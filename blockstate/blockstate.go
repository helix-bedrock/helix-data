// Package blockstate reads the block states BDS saved for the bds-dump
// add-on, writes them as data files, and compares them with other palettes.
package blockstate

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

// Dump is what one run of the add-on saved.
type Dump struct {
	// States lists every distinct state, sorted by Key.
	States []VersionedState
	// Defaults maps each block name to the properties of its default state.
	Defaults map[string]map[string]any
}

// Structure identifier prefixes the add-on saves under: every combination
// and default permutations.
const (
	allPrefix     = "helixdump:p"
	defaultPrefix = "helixdump:d"
)

// chunkOnly lists blocks that cannot be stored in a structure; the add-on
// leaves one in the world and FromWorld reads it from the chunk.
var chunkOnly = []string{"minecraft:structure_void"}

// FromWorld reads the structures the add-on saved in a BDS world's db folder.
func FromWorld(dbDir string) (*Dump, error) {
	all, err := ReadStructures(dbDir, allPrefix)
	if err != nil {
		return nil, err
	}
	fromChunks, err := ReadChunkStates(dbDir, chunkOnly)
	if err != nil {
		return nil, err
	}
	for _, name := range chunkOnly {
		if len(fromChunks[name]) == 0 {
			return nil, fmt.Errorf("blockstate: %s not found in any chunk of %s", name, dbDir)
		}
		all["chunk:"+name] = fromChunks[name]
	}
	defs, err := ReadStructures(dbDir, defaultPrefix)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 || len(defs) == 0 {
		return nil, fmt.Errorf("blockstate: %s has no helixdump structures; did the dump finish and the server stop cleanly?", dbDir)
	}
	d := &Dump{Defaults: map[string]map[string]any{}}
	seen := map[string]bool{}
	for _, list := range all {
		for _, s := range list {
			if k := s.Key(); !seen[k] {
				seen[k] = true
				d.States = append(d.States, s)
			}
		}
	}
	slices.SortFunc(d.States, func(a, b VersionedState) int { return strings.Compare(a.Key(), b.Key()) })
	for _, name := range chunkOnly {
		defs["chunk:"+name] = fromChunks[name][:1]
	}
	for _, list := range defs {
		for _, s := range list {
			if old, ok := d.Defaults[s.Name]; ok && (State{s.Name, old}).Key() != s.Key() {
				return nil, fmt.Errorf("blockstate: two default states for %s", s.Name)
			}
			d.Defaults[s.Name] = s.Properties
		}
	}
	return d, nil
}

// WriteNBT writes states as little-endian NBT compounds one after another,
// each {name, states, version}: the canonical_block_states.nbt layout pmmp
// and Dragonfly read.
func WriteNBT(w io.Writer, states []VersionedState) error {
	for _, s := range states {
		props := s.Properties
		if props == nil {
			props = map[string]any{}
		}
		b, err := nbt.MarshalEncoding(map[string]any{
			"name":    s.Name,
			"states":  props,
			"version": s.Version,
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

// Plain returns the states without version numbers, for Compare.
func (d *Dump) Plain() []State {
	out := make([]State, len(d.States))
	for i, s := range d.States {
		out[i] = s.State
	}
	return out
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
