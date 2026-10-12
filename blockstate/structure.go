package blockstate

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/helix-bedrock/protocol/nbt"
	"github.com/helix-bedrock/world/leveldb"
	"github.com/helix-bedrock/world/leveldb/opt"
	"github.com/helix-bedrock/world/leveldb/util"
)

// structureKeyPrefix is the LevelDB key prefix BDS uses for structures saved
// to the world.
const structureKeyPrefix = "structuretemplate_"

// VersionedState is a state as BDS saved it, with its block version number.
type VersionedState struct {
	State
	Version int32
}

// ReadStructures opens the world database in dbDir read-only and returns the
// block palettes of every structure whose identifier starts with idPrefix
// (for example "helixdump:p"), keyed by identifier.
func ReadStructures(dbDir, idPrefix string) (map[string][]VersionedState, error) {
	db, err := leveldb.OpenFile(dbDir, &opt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("blockstate: open %s: %w", dbDir, err)
	}
	defer func() { _ = db.Close() }()

	out := map[string][]VersionedState{}
	it := db.NewIterator(util.BytesPrefix([]byte(structureKeyPrefix+idPrefix)), nil)
	defer it.Release()
	for it.Next() {
		id := strings.TrimPrefix(string(it.Key()), structureKeyPrefix)
		// A map, not a struct: the decoder rejects fields a struct lacks, and
		// structures have many we do not need.
		var root map[string]any
		if err := nbt.NewDecoderWithEncoding(bytes.NewReader(it.Value()), nbt.LittleEndian).Decode(&root); err != nil {
			return nil, fmt.Errorf("blockstate: structure %s: %w", id, err)
		}
		structure, _ := root["structure"].(map[string]any)
		palettes, _ := structure["palette"].(map[string]any)
		def, _ := palettes["default"].(map[string]any)
		list, ok := def["block_palette"].([]any)
		if !ok {
			return nil, fmt.Errorf("blockstate: structure %s has no default block palette", id)
		}
		for i, e := range list {
			m, _ := e.(map[string]any)
			name, ok := m["name"].(string)
			if !ok {
				return nil, fmt.Errorf("blockstate: structure %s palette entry %d has no name", id, i)
			}
			props, _ := m["states"].(map[string]any)
			if props == nil {
				props = map[string]any{}
			}
			version, _ := m["version"].(int32)
			out[id] = append(out[id], VersionedState{State{name, props}, version})
		}
	}
	if err := it.Error(); err != nil {
		return nil, fmt.Errorf("blockstate: read %s: %w", dbDir, err)
	}
	return out, nil
}
