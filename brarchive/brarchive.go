// Package brarchive reads the .brarchive files BDS packs its behaviour pack
// folders into (entities, recipes, spawn rules and so on).
//
// The layout was read from the BDS 1.26.52.3 files: a 16-byte header (an
// 8-byte magic, the file count and a version, little-endian uint32s), then
// one 256-byte entry per file (a name length byte, the name padded to 247
// bytes, then the data offset and size as uint32s), then the data, with
// offsets counted from the end of the entries.
package brarchive

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// File is one file in an archive.
type File struct {
	Name string
	Data []byte
}

const (
	headerSize = 16
	entrySize  = 256
	maxName    = 247
)

// Read returns the files in an archive. The data slices point into b.
func Read(b []byte) ([]File, error) {
	if len(b) < headerSize {
		return nil, errors.New("brarchive: file shorter than its header")
	}
	n := binary.LittleEndian.Uint32(b[8:])
	if uint64(n)*entrySize > uint64(len(b)-headerSize) {
		return nil, fmt.Errorf("brarchive: %d entries do not fit in %d bytes", n, len(b))
	}
	base := headerSize + int(n)*entrySize
	files := make([]File, 0, n)
	for i := range int(n) {
		e := b[headerSize+i*entrySize : headerSize+(i+1)*entrySize]
		l := int(e[0])
		if l > maxName {
			return nil, fmt.Errorf("brarchive: entry %d: name length %d", i, l)
		}
		off, size := binary.LittleEndian.Uint32(e[248:]), binary.LittleEndian.Uint32(e[252:])
		start, end := uint64(base)+uint64(off), uint64(base)+uint64(off)+uint64(size)
		if end > uint64(len(b)) {
			return nil, fmt.Errorf("brarchive: entry %d (%s) ends past the file", i, e[1:1+l])
		}
		files = append(files, File{Name: string(e[1 : 1+l]), Data: b[start:end]})
	}
	return files, nil
}
