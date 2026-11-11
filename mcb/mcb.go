// Package mcb reads the binary form ("\x7fMCB") some vanilla item definitions take in
// BDS's .brarchive files, and turns them into the JSON definition they stand for.
//
// The layout, read from BDS 1.26.60's own files: the magic, a uint16 version (1), a
// uint16 and a uint32 BDS uses for its own versioning, then the root ("minecraft:item"),
// the identifier, the menu category, its group and whether it is hidden in commands,
// and the components. Each component is the FNV-1a 32-bit hash of its name followed by
// its fields in a fixed order, so only components whose fields are known here can be
// read; any other makes Decode fail rather than drop data. Strings are a varuint32
// length and the bytes, numbers little-endian.
package mcb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
)

// Magic starts a definition in binary form.
var Magic = []byte{0x7f, 'M', 'C', 'B'}

// component reads one component's fields.
type component func(r *reader) any

// components are the item components whose binary fields are known, by name.
var components = map[string]component{
	"minecraft:display_name": func(r *reader) any { return map[string]any{"value": r.str()} },
	"minecraft:icon": func(r *reader) any {
		textures := map[string]any{}
		for n := r.count(); n > 0 && r.err == nil; n-- {
			k := r.str()
			textures[k] = r.str()
		}
		return map[string]any{"textures": textures}
	},
	"minecraft:max_stack_size": func(r *reader) any { return r.u16() },
	"minecraft:hand_equipped":  func(r *reader) any { return r.boolean() },
	"minecraft:cooldown": func(r *reader) any {
		return map[string]any{"category": r.str(), "duration": r.f32(), "type": r.str()}
	},
	"minecraft:projectile": func(r *reader) any {
		return map[string]any{"minimum_critical_power": r.f32(), "projectile_entity": r.str()}
	},
	"minecraft:throwable": func(r *reader) any {
		return map[string]any{
			"do_swing_animation": r.boolean(), "max_draw_duration": r.f32(), "min_draw_duration": r.f32(),
			"launch_power_scale": r.f32(), "max_launch_power": r.f32(), "scale_power_by_draw_duration": r.boolean(),
		}
	},
	"minecraft:block_placer": func(r *reader) any {
		block := r.str()
		states := r.descriptorStates()
		replace := r.boolean()
		var useOn []any
		for n := r.count(); n > 0 && r.err == nil; n-- {
			useOn = append(useOn, r.descriptor())
		}
		b := any(block)
		if len(states) > 0 {
			b = map[string]any{"name": block, "states": states}
		}
		out := map[string]any{"block": b, "replace_block_item": replace}
		if len(useOn) > 0 {
			out["use_on"] = useOn
		}
		return out
	},
	"minecraft:compostable": func(r *reader) any { return map[string]any{"composting_chance": r.u8()} },
	"minecraft:tags": func(r *reader) any {
		tags := []any{}
		for n := r.count(); n > 0 && r.err == nil; n-- {
			tags = append(tags, r.str())
		}
		return map[string]any{"tags": tags}
	},
	"minecraft:bundle_interaction": func(r *reader) any { return map[string]any{"num_viewable_slots": r.u32()} },
	"minecraft:storage_item": func(r *reader) any {
		out := map[string]any{"max_slots": r.u32(), "allow_nested_storage_items": r.boolean()}
		if banned := r.descriptors(); len(banned) > 0 {
			out["banned_items"] = banned
		}
		if allowed := r.descriptors(); len(allowed) > 0 {
			out["allowed_items"] = allowed
		}
		return out
	},
	"minecraft:storage_weight_limit": func(r *reader) any {
		return map[string]any{"max_weight_limit": r.u32()}
	},
	"minecraft:storage_weight_modifier": func(r *reader) any {
		return map[string]any{"weight_in_storage_item": r.u32()}
	},
}

// byHash is components keyed by the hash BDS writes for them.
var byHash = func() map[uint32]string {
	m := make(map[uint32]string, len(components))
	for name := range components {
		m[hash(name)] = name
	}
	return m
}()

// hash is the FNV-1a 32-bit hash BDS names fields by.
func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// Decode turns a binary item definition into its JSON definition:
// {"minecraft:item": {"description": {...}, "components": {...}}}.
func Decode(b []byte) (json.RawMessage, error) {
	if !bytes.HasPrefix(b, Magic) {
		return nil, errors.New("mcb: not a binary definition")
	}
	r := &reader{r: bytes.NewReader(b[len(Magic):])}
	if v := r.u16(); v != 1 {
		return nil, fmt.Errorf("mcb: version %d", v)
	}
	r.u16()
	r.u32()
	root := r.str()
	desc := map[string]any{"identifier": r.str()}
	category := map[string]any{"category": r.str()}
	if group := r.str(); group != "" {
		category["group"] = group
	}
	if r.boolean() {
		category["is_hidden_in_commands"] = true
	}
	desc["menu_category"] = category
	comps := map[string]any{}
	for n := r.u32(); n > 0 && r.err == nil; n-- {
		h := r.u32()
		name, ok := byHash[h]
		if !ok {
			return nil, fmt.Errorf("mcb: %v: unknown component hash %#08x", desc["identifier"], h)
		}
		comps[name] = components[name](r)
	}
	if r.err != nil {
		return nil, fmt.Errorf("mcb: %v: %w", desc["identifier"], r.err)
	}
	if r.r.Len() != 0 {
		return nil, fmt.Errorf("mcb: %v: %d bytes left over", desc["identifier"], r.r.Len())
	}
	return json.Marshal(map[string]any{root: map[string]any{"description": desc, "components": comps}})
}

// reader reads fields, keeping the first error.
type reader struct {
	r   *bytes.Reader
	err error
}

func (r *reader) read(n int) []byte {
	if r.err != nil {
		return make([]byte, n)
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r.r, b); err != nil {
		r.err = err
	}
	return b
}

func (r *reader) u8() uint8     { return r.read(1)[0] }
func (r *reader) boolean() bool { return r.u8() != 0 }
func (r *reader) u16() uint16   { return binary.LittleEndian.Uint16(r.read(2)) }
func (r *reader) u32() uint32   { return binary.LittleEndian.Uint32(r.read(4)) }
func (r *reader) f32() float32  { return math.Float32frombits(r.u32()) }
func (r *reader) count() uint32 { return r.varuint() }
func (r *reader) str() string   { return string(r.read(int(r.varuint()))) }
func (r *reader) varuint() uint32 {
	var v uint32
	for shift := 0; shift < 35; shift += 7 {
		b := r.u8()
		if r.err != nil {
			return 0
		}
		v |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			return v
		}
	}
	r.err = errors.New("varuint too long")
	return 0
}

// descriptor reads an item or block descriptor: a map of string fields such as "name".
func (r *reader) descriptor() map[string]any {
	out := map[string]any{}
	for n := r.count(); n > 0 && r.err == nil; n-- {
		k := r.str()
		out[k] = r.str()
	}
	return out
}

// descriptors reads a list of descriptors, each written as its one name.
func (r *reader) descriptors() []any {
	var out []any
	for n := r.count(); n > 0 && r.err == nil; n-- {
		d := r.descriptor()
		if name, ok := d["name"]; ok && len(d) == 1 {
			out = append(out, name)
			continue
		}
		out = append(out, d)
	}
	return out
}

// descriptorStates reads a block descriptor's states, written as string pairs.
func (r *reader) descriptorStates() map[string]any {
	out := map[string]any{}
	for n := r.count(); n > 0 && r.err == nil; n-- {
		k := r.str()
		out[k] = r.str()
	}
	return out
}
