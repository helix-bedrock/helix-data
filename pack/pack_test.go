package pack

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// archive builds a .brarchive in the layout brarchive.Read reads.
func archive(files map[string]string) []byte {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	size := 0
	for _, d := range files {
		size += len(d)
	}
	head := make([]byte, 16+256*len(names), 16+256*len(names)+size)
	binary.LittleEndian.PutUint32(head[8:], uint32(len(names)))
	data := make([]byte, 0, size)
	for i, n := range names {
		e := head[16+i*256:]
		e[0] = byte(len(n))
		copy(e[1:], n)
		binary.LittleEndian.PutUint32(e[248:], uint32(len(data)))
		binary.LittleEndian.PutUint32(e[252:], uint32(len(files[n])))
		data = append(data, files[n]...)
	}
	return append(head, data...)
}

func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStack(t *testing.T) {
	bds := t.TempDir()
	bp := filepath.Join(bds, "behavior_packs")
	write(t, filepath.Join(bds, "definitions", "features", "base.json"), []byte("def"))
	write(t, filepath.Join(bds, "definitions", "persona", "x.json"), []byte("skip"))
	write(t, filepath.Join(bp, "vanilla", "__brarchive", "features.brarchive"), archive(map[string]string{"base.json": "old", "a.json": "a0"}))
	write(t, filepath.Join(bp, "vanilla", "structures", "s.nbt"), []byte("nbt0"))
	write(t, filepath.Join(bp, "vanilla", "manifest.json"), []byte("{}"))
	// Archives list structures as empty entries next to the real file.
	write(t, filepath.Join(bp, "vanilla_1.10", "__brarchive", "structures.brarchive"), archive(map[string]string{"s.nbt": ""}))
	write(t, filepath.Join(bp, "vanilla_1.10", "structures", "s.nbt"), []byte("nbt1"))
	// 1.9 sorts before 1.10 by number, not by name.
	write(t, filepath.Join(bp, "vanilla_1.9", "__brarchive", "features.brarchive"), archive(map[string]string{"a.json": "a9"}))
	write(t, filepath.Join(bp, "vanilla_1.10", "__brarchive", "features.brarchive"), archive(map[string]string{"a.json": "a10"}))
	write(t, filepath.Join(bp, "vanilla_1.11", "__brarchive", "structures.brarchive"), archive(map[string]string{"s.nbt": ""}))

	got, err := Stack(bds, []string{"features", "structures"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]File{
		"features/base.json": {Data: []byte("old"), From: "vanilla"},
		"features/a.json":    {Data: []byte("a10"), From: "vanilla_1.10"},
		"structures/s.nbt":   {Data: []byte("nbt1"), From: "vanilla_1.10"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d files, want %d: %v", len(got), len(want), got)
	}
	for k, w := range want {
		g := got[k]
		if string(g.Data) != string(w.Data) || g.From != w.From {
			t.Errorf("%s: got %q from %s, want %q from %s", k, g.Data, g.From, w.Data, w.From)
		}
	}
}
