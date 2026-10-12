package blockstate

import (
	"bytes"
	"testing"

	"github.com/helix-bedrock/protocol/nbt"
	"github.com/helix-bedrock/world/leveldb"
)

// putStructure saves a structure with the given palette the way BDS does.
func putStructure(t *testing.T, db *leveldb.DB, id string, palette []map[string]any) {
	t.Helper()
	list := make([]any, len(palette))
	for i, p := range palette {
		list[i] = p
	}
	b, err := nbt.MarshalEncoding(map[string]any{
		"format_version": int32(1),
		"size":           []int32{1, 1, 1},
		"structure": map[string]any{
			"palette": map[string]any{
				"default": map[string]any{"block_palette": list},
			},
		},
	}, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte(structureKeyPrefix+id), b, nil); err != nil {
		t.Fatal(err)
	}
}

// putSubChunk saves a one-layer sub chunk whose only block is st.
func putSubChunk(t *testing.T, db *leveldb.DB, st map[string]any) {
	t.Helper()
	b, err := nbt.MarshalEncoding(st, nbt.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	key := []byte{0, 0, 0, 0, 0, 0, 0, 0, subChunkTag, 4}
	// Version 9, one layer, Y 4, then a storage with 0 bits per block: no
	// index words and an implicit palette of one entry.
	val := append([]byte{9, 1, 4, 0}, b...)
	if err := db.Put(key, val, nil); err != nil {
		t.Fatal(err)
	}
}

func state(name string, version int32, props map[string]any) map[string]any {
	return map[string]any{"name": name, "states": props, "version": version}
}

func testWorld(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	lever := func(open uint8) map[string]any {
		return state("minecraft:lever", 18168865, map[string]any{"open_bit": open, "lever_direction": "down_east_west"})
	}
	putStructure(t, db, "helixdump:p1", []map[string]any{lever(0), lever(1), state("minecraft:stone", 18168865, map[string]any{})})
	putStructure(t, db, "helixdump:p2", []map[string]any{lever(1)}) // duplicate across structures
	putStructure(t, db, "helixdump:d1", []map[string]any{lever(0), state("minecraft:stone", 18168865, map[string]any{})})
	putSubChunk(t, db, state("minecraft:structure_void", 18168865, map[string]any{}))
	putStructure(t, db, "other:x", []map[string]any{state("minecraft:dirt", 1, map[string]any{})})
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestFromWorld(t *testing.T) {
	d, err := FromWorld(testWorld(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.States) != 4 || len(d.Defaults) != 3 {
		t.Fatalf("got %d states and %d defaults, want 4 and 3", len(d.States), len(d.Defaults))
	}
	if v := d.Defaults["minecraft:lever"]["open_bit"]; v != uint8(0) {
		t.Fatalf("open_bit default = %#v, want uint8(0)", v)
	}
	for _, s := range d.States {
		if s.Version != 18168865 || s.Name == "minecraft:dirt" {
			t.Fatalf("unexpected state %+v", s)
		}
	}
}

func TestFromWorldEmpty(t *testing.T) {
	dir := t.TempDir()
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := FromWorld(dir); err == nil {
		t.Fatal("no error for a world without a dump")
	}
}

func TestWriteNBT(t *testing.T) {
	d, err := FromWorld(testWorld(t))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteNBT(&buf, d.States); err != nil {
		t.Fatal(err)
	}
	dec := nbt.NewDecoderWithEncoding(&buf, nbt.LittleEndian)
	var got []State
	for buf.Len() > 0 {
		var m struct {
			Name    string         `nbt:"name"`
			States  map[string]any `nbt:"states"`
			Version int32          `nbt:"version"`
		}
		if err := dec.Decode(&m); err != nil {
			t.Fatal(err)
		}
		if m.Version != 18168865 {
			t.Fatalf("version %d", m.Version)
		}
		got = append(got, State{Name: m.Name, Properties: m.States})
	}
	if diff := Compare(d.Plain(), got); len(diff.OnlyA)+len(diff.OnlyB) != 0 || diff.Same != 4 {
		t.Fatalf("round trip differs: %+v", diff)
	}
}

func TestKeyIgnoresOrder(t *testing.T) {
	a := State{"b", map[string]any{"x": int32(1), "y": "z"}}
	b := State{"b", map[string]any{"y": "z", "x": int32(1)}}
	c := State{"b", map[string]any{"y": "z", "x": uint8(1)}}
	if a.Key() != b.Key() || a.Key() == c.Key() {
		t.Fatal("Key must ignore map order and keep value types apart")
	}
}

func TestSubChunkPalettesCorrupt(t *testing.T) {
	for _, data := range [][]byte{
		{9},                    // too short
		{9, 1, 0, 7 << 1},      // invalid bits per block
		{9, 1, 0, 1 << 1, 0},   // index words cut off
		{9, 2, 0, 0, 10, 0, 0}, // broken NBT
	} {
		if _, err := subChunkPalettes(data); err == nil {
			t.Fatalf("no error for %v", data)
		}
	}
}

func TestNetworkHash(t *testing.T) {
	// Network IDs BDS 1.26.52.3 sent to a client for these states.
	for _, c := range []struct {
		s    State
		want uint32
	}{
		{State{"minecraft:air", map[string]any{}}, 3690217760},
		{State{"minecraft:cobblestone", map[string]any{}}, 1741778478},
		{State{"minecraft:fire", map[string]any{"age": int32(0)}}, 2464443264},
		{State{"minecraft:bubble_column", map[string]any{"drag_down": uint8(1)}}, 3430338787},
		{State{"minecraft:campfire", map[string]any{"extinguished": uint8(1), "minecraft:cardinal_direction": "west"}}, 2992575859},
	} {
		got, err := NetworkHash(c.s)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%s: got %d, want %d", c.s.Key(), got, c.want)
		}
	}
	if _, err := NetworkHash(State{"x", map[string]any{"f": 1.5}}); err == nil {
		t.Fatal("no error for an unsupported value type")
	}
}
