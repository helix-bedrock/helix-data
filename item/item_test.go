package item

import (
	"bytes"
	"strings"
	"testing"
)

func TestFromLogRoundTrip(t *testing.T) {
	log := `[INFO] other line
[2026 INFO] [Scripting] HELIXDUMP-ITEM {"name":"minecraft:stone","maxStack":64}
[2026 INFO] [Scripting] HELIXDUMP-ITEM {"name":"minecraft:bed","maxStack":1}
`
	items, err := FromLog(strings.NewReader(log))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "minecraft:bed" || items[1].MaxStack != 64 {
		t.Fatalf("got %+v", items)
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, items); err != nil {
		t.Fatal(err)
	}
	back, err := ReadJSON(&buf)
	if err != nil || len(back) != 2 || back[0] != items[0] {
		t.Fatalf("round trip: %+v, %v", back, err)
	}
}

func TestFromLogRejects(t *testing.T) {
	for _, log := range []string{
		"",
		`HELIXDUMP-ITEM {"name":"minecraft:a","maxStack":0}`,
		`HELIXDUMP-ITEM {"name":"","maxStack":1}`,
		"HELIXDUMP-ITEM {\"name\":\"minecraft:a\",\"maxStack\":1}\nHELIXDUMP-ITEM {\"name\":\"minecraft:a\",\"maxStack\":1}",
		`HELIXDUMP-ITEM {broken`,
	} {
		if _, err := FromLog(strings.NewReader(log)); err == nil {
			t.Errorf("%q: no error", log)
		}
	}
}
