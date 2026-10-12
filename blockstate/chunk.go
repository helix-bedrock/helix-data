package blockstate

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/helix-bedrock/protocol/nbt"
	"github.com/helix-bedrock/world/leveldb"
	"github.com/helix-bedrock/world/leveldb/opt"
)

// subChunkTag is the key byte of a sub chunk record: x, z, optional
// dimension, then this tag and the sub chunk's Y index.
const subChunkTag = '/'

// ReadChunkStates returns the block states named in names that appear in the
// palettes of any sub chunk in the world database in dbDir. Only palettes are
// read, not block positions.
func ReadChunkStates(dbDir string, names []string) (map[string][]VersionedState, error) {
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}
	db, err := leveldb.OpenFile(dbDir, &opt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("blockstate: open %s: %w", dbDir, err)
	}
	defer func() { _ = db.Close() }()

	out := map[string][]VersionedState{}
	seen := map[string]bool{}
	it := db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		k := it.Key()
		isSub := (len(k) == 10 && k[8] == subChunkTag) || (len(k) == 14 && k[12] == subChunkTag)
		if !isSub {
			continue
		}
		states, err := subChunkPalettes(it.Value())
		if err != nil {
			return nil, fmt.Errorf("blockstate: sub chunk %x: %w", k, err)
		}
		for _, s := range states {
			if want[s.Name] && !seen[s.Key()] {
				seen[s.Key()] = true
				out[s.Name] = append(out[s.Name], s)
			}
		}
	}
	if err := it.Error(); err != nil {
		return nil, fmt.Errorf("blockstate: read %s: %w", dbDir, err)
	}
	return out, nil
}

// subChunkPalettes returns the palette entries of every layer in a sub chunk
// stored in the format versions 8 and 9 that current Bedrock writes.
func subChunkPalettes(data []byte) ([]VersionedState, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("sub chunk too short")
	}
	var layers int
	r := bytes.NewReader(data)
	switch v, _ := r.ReadByte(); v {
	case 8, 9:
		n, _ := r.ReadByte()
		layers = int(n)
		if v == 9 {
			if _, err := r.ReadByte(); err != nil {
				return nil, fmt.Errorf("sub chunk Y: %w", err)
			}
		}
	default:
		// Older formats predate block states in palettes; nothing to read.
		return nil, nil
	}
	var out []VersionedState
	for l := 0; l < layers; l++ {
		h, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("layer %d header: %w", l, err)
		}
		bits := int(h >> 1)
		count := uint32(1)
		if bits != 0 {
			switch bits {
			case 1, 2, 3, 4, 5, 6, 8, 16:
			default:
				return nil, fmt.Errorf("layer %d: invalid bits per block %d", l, bits)
			}
			perWord := 32 / bits
			words := (4096 + perWord - 1) / perWord
			if _, err := r.Seek(int64(words*4), 1); err != nil {
				return nil, fmt.Errorf("layer %d: %w", l, err)
			}
			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
				return nil, fmt.Errorf("layer %d palette size: %w", l, err)
			}
			if count > 4096 {
				return nil, fmt.Errorf("layer %d: palette size %d over 4096", l, count)
			}
		}
		dec := nbt.NewDecoderWithEncoding(r, nbt.LittleEndian)
		for i := uint32(0); i < count; i++ {
			var m map[string]any
			if err := dec.Decode(&m); err != nil {
				return nil, fmt.Errorf("layer %d palette entry %d: %w", l, i, err)
			}
			name, _ := m["name"].(string)
			props, _ := m["states"].(map[string]any)
			if props == nil {
				props = map[string]any{}
			}
			version, _ := m["version"].(int32)
			out = append(out, VersionedState{State{name, props}, version})
		}
	}
	return out, nil
}
