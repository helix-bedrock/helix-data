package brarchive

import (
	"encoding/binary"
	"testing"
)

func archive(files ...File) []byte {
	size := 0
	for _, f := range files {
		size += len(f.Data)
	}
	b := make([]byte, headerSize+entrySize*len(files), headerSize+entrySize*len(files)+size)
	binary.LittleEndian.PutUint32(b[8:], uint32(len(files)))
	data := make([]byte, 0, size)
	for i, f := range files {
		e := b[headerSize+i*entrySize:]
		e[0] = byte(len(f.Name))
		copy(e[1:], f.Name)
		binary.LittleEndian.PutUint32(e[248:], uint32(len(data)))
		binary.LittleEndian.PutUint32(e[252:], uint32(len(f.Data)))
		data = append(data, f.Data...)
	}
	return append(b, data...)
}

func TestRead(t *testing.T) {
	files, err := Read(archive(File{"a.json", []byte("{}")}, File{"b.json", []byte(`{"x":1}`)}))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[1].Name != "b.json" || string(files[1].Data) != `{"x":1}` {
		t.Fatalf("got %+v", files)
	}
}

func TestReadRejectsBadOffsets(t *testing.T) {
	b := archive(File{"a.json", []byte("{}")})
	binary.LittleEndian.PutUint32(b[headerSize+248:], 1<<31)
	if _, err := Read(b); err == nil {
		t.Fatal("no error for an offset past the end")
	}
}

func FuzzRead(f *testing.F) {
	f.Add(archive(File{"a.json", []byte("{}")}))
	f.Add([]byte{})
	f.Fuzz(func(_ *testing.T, b []byte) { _, _ = Read(b) })
}
