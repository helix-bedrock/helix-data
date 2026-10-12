// Package blockstate reads the block state dump printed by the bds-dump
// add-on, writes it as data files, and compares it with other palettes.
package blockstate

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"github.com/helix-bedrock/protocol/nbt"
)

// State is one block state: a block name and its property values. Values are
// int32, string or uint8 (Bedrock stores boolean properties as bytes).
type State struct {
	Name       string
	Properties map[string]any
}

// Key returns a string that is equal for equal states, whatever the map order.
func (s State) Key() string {
	names := make([]string, 0, len(s.Properties))
	for n := range s.Properties {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString(s.Name)
	for _, n := range names {
		fmt.Fprintf(&b, ";%s=%T:%v", n, s.Properties[n], s.Properties[n])
	}
	return b.String()
}

// Dump is everything one run of the add-on printed.
type Dump struct {
	// States lists every state in the order BDS printed them.
	States []State
	// Defaults maps each block name to its default property values.
	Defaults map[string]map[string]any
	// Skipped lists combinations BDS refused, with its error.
	Skipped []string
}

const (
	markState   = "HELIXDUMP "
	markDefault = "HELIXDUMP-DEFAULT "
	markSkip    = "HELIXDUMP-SKIP "
	markEnd     = "HELIXDUMP-END "
)

type line struct {
	Name   string         `json:"name"`
	States map[string]any `json:"states"`
	Error  string         `json:"error"`
}

// ReadDump reads a BDS console log. Lines without a HELIXDUMP marker are
// ignored, so the whole log can be passed in. It fails if the log does not
// contain the end marker, which means the dump was cut off.
func ReadDump(r io.Reader) (*Dump, error) {
	d := &Dump{Defaults: map[string]map[string]any{}}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	ended := false
	n := 0
	for sc.Scan() {
		n++
		text := sc.Text()
		mark, rest, ok := cut(text)
		if !ok {
			continue
		}
		if mark == markEnd {
			ended = true
			continue
		}
		var l line
		if err := json.Unmarshal([]byte(rest), &l); err != nil {
			return nil, fmt.Errorf("blockstate: line %d: %w", n, err)
		}
		props, err := convert(l.States)
		if err != nil {
			return nil, fmt.Errorf("blockstate: line %d: %s: %w", n, l.Name, err)
		}
		switch mark {
		case markState:
			d.States = append(d.States, State{Name: l.Name, Properties: props})
		case markDefault:
			d.Defaults[l.Name] = props
		case markSkip:
			d.Skipped = append(d.Skipped, fmt.Sprintf("%s %v: %s", l.Name, l.States, l.Error))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("blockstate: read log: %w", err)
	}
	if !ended {
		return nil, errors.New("blockstate: no HELIXDUMP-END line; the dump did not finish")
	}
	return d, nil
}

// cut finds the marker in a log line. BDS prefixes script output with a
// timestamp and "[Scripting]", so the marker can be anywhere.
func cut(text string) (mark, rest string, ok bool) {
	i := strings.Index(text, "HELIXDUMP")
	if i < 0 {
		return "", "", false
	}
	text = text[i:]
	for _, m := range []string{markDefault, markSkip, markEnd, markState} {
		if r, found := strings.CutPrefix(text, m); found {
			return m, r, true
		}
	}
	return "", "", false
}

// convert turns JSON values into the NBT types Bedrock uses.
func convert(in map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch v := v.(type) {
		case bool:
			if v {
				out[k] = uint8(1)
			} else {
				out[k] = uint8(0)
			}
		case float64:
			if v != float64(int32(v)) {
				return nil, fmt.Errorf("property %s: %v is not an int32", k, v)
			}
			out[k] = int32(v)
		case string:
			out[k] = v
		default:
			return nil, fmt.Errorf("property %s: unsupported type %T", k, v)
		}
	}
	return out, nil
}

// WriteNBT writes states as little-endian NBT compounds one after another,
// each {name, states, version}: the canonical_block_states.nbt layout pmmp
// and Dragonfly read.
func WriteNBT(w io.Writer, states []State, version int32) error {
	for _, s := range states {
		props := s.Properties
		if props == nil {
			props = map[string]any{}
		}
		b, err := nbt.MarshalEncoding(map[string]any{
			"name":    s.Name,
			"states":  props,
			"version": version,
		}, nbt.LittleEndian)
		if err != nil {
			return fmt.Errorf("blockstate: encode %s: %w", s.Name, err)
		}
		if _, err := w.Write(b); err != nil {
			return fmt.Errorf("blockstate: write: %w", err)
		}
	}
	return nil
}

// ReadPalette reads a gzipped big-endian palette with a "blocks" list, the
// block_palette.nbt layout CloudburstMC/Data publishes.
func ReadPalette(r io.Reader) ([]State, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("blockstate: palette: %w", err)
	}
	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("blockstate: palette: %w", err)
	}
	var root struct {
		Blocks []map[string]any `nbt:"blocks"`
	}
	if err := nbt.UnmarshalEncoding(data, &root, nbt.BigEndian); err != nil {
		return nil, fmt.Errorf("blockstate: palette: %w", err)
	}
	out := make([]State, 0, len(root.Blocks))
	for i, m := range root.Blocks {
		name, ok := m["name"].(string)
		if !ok {
			return nil, fmt.Errorf("blockstate: palette entry %d has no name", i)
		}
		props, _ := m["states"].(map[string]any)
		if props == nil {
			props = map[string]any{}
		}
		out = append(out, State{Name: name, Properties: props})
	}
	return out, nil
}

// Diff is the result of Compare.
type Diff struct {
	// OnlyA and OnlyB list state keys found in one side only, sorted.
	OnlyA, OnlyB []string
	// Same counts states found in both.
	Same int
}

// Compare compares two state lists as sets.
func Compare(a, b []State) Diff {
	inA := make(map[string]bool, len(a))
	for _, s := range a {
		inA[s.Key()] = true
	}
	inB := make(map[string]bool, len(b))
	for _, s := range b {
		inB[s.Key()] = true
	}
	var d Diff
	for k := range inA {
		if inB[k] {
			d.Same++
		} else {
			d.OnlyA = append(d.OnlyA, k)
		}
	}
	for k := range inB {
		if !inA[k] {
			d.OnlyB = append(d.OnlyB, k)
		}
	}
	slices.Sort(d.OnlyA)
	slices.Sort(d.OnlyB)
	return d
}

// DefaultsJSON returns the defaults as indented JSON with sorted keys.
func (d *Dump) DefaultsJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d.Defaults); err != nil {
		return nil, fmt.Errorf("blockstate: defaults: %w", err)
	}
	return buf.Bytes(), nil
}
