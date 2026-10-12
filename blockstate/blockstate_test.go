package blockstate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/helix-bedrock/protocol/nbt"
)

const sampleLog = `NO LOG FILE! - [2026-10-11 12:00:00 INFO] Server started.
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP-BEGIN {"blocks":2}
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP-DEFAULT {"name":"minecraft:stone","states":{}}
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP {"name":"minecraft:stone","states":{}}
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP-DEFAULT {"name":"minecraft:lever","states":{"open_bit":false,"lever_direction":"down_east_west"}}
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP {"name":"minecraft:lever","states":{"open_bit":false,"lever_direction":"down_east_west"}}
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP {"name":"minecraft:lever","states":{"open_bit":true,"lever_direction":"down_east_west"}}
[2026-10-11 12:00:01:000 INFO] [Scripting] HELIXDUMP-SKIP {"name":"minecraft:x","states":{"age":9},"error":"bad"}
[2026-10-11 12:00:02:000 INFO] [Scripting] HELIXDUMP-END {"blocks":2,"states":3,"skipped":1}
`

func TestReadDump(t *testing.T) {
	d, err := ReadDump(strings.NewReader(sampleLog))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.States) != 3 || len(d.Defaults) != 2 || len(d.Skipped) != 1 {
		t.Fatalf("got %d states, %d defaults, %d skipped", len(d.States), len(d.Defaults), len(d.Skipped))
	}
	if v := d.Defaults["minecraft:lever"]["open_bit"]; v != uint8(0) {
		t.Fatalf("open_bit default = %#v, want uint8(0)", v)
	}
}

func TestReadDumpUnfinished(t *testing.T) {
	cut, _, _ := strings.Cut(sampleLog, "HELIXDUMP-END")
	if _, err := ReadDump(strings.NewReader(cut)); err == nil {
		t.Fatal("no error for a dump without an end line")
	}
}

func TestReadDumpBadValue(t *testing.T) {
	log := `HELIXDUMP {"name":"a","states":{"x":1.5}}` + "\nHELIXDUMP-END {}\n"
	if _, err := ReadDump(strings.NewReader(log)); err == nil {
		t.Fatal("no error for a non-integer number")
	}
}

func TestWriteNBT(t *testing.T) {
	d, err := ReadDump(strings.NewReader(sampleLog))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteNBT(&buf, d.States, 18168865); err != nil {
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
	if diff := Compare(d.States, got); len(diff.OnlyA)+len(diff.OnlyB) != 0 || diff.Same != 3 {
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
