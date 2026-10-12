// Package data gives Go programs the generated game data in data/, so the
// engine can load it without a copy of this repo on disk.
package data

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/helix-bedrock/helix-data/blockstate"
	"github.com/helix-bedrock/helix-data/item"
)

// Version is the Minecraft version the data is from.
const Version = "1.26.52"

//go:embed 1.26.52/block_states.nbt
var blockStates []byte

// BlockStates returns every block state of [Version] as BDS saved it, in the
// order of data/1.26.52/block_states.nbt. Each call decodes a fresh copy.
func BlockStates() ([]blockstate.VersionedState, error) {
	states, err := blockstate.ReadNBT(bytes.NewReader(blockStates))
	if err != nil {
		return nil, fmt.Errorf("data: block states: %w", err)
	}
	return states, nil
}

//go:embed 1.26.52/items.json
var items []byte

// Items returns every item type of [Version] with its stack size, as BDS
// reports it, sorted by name. Each call decodes a fresh copy.
func Items() ([]item.Item, error) {
	list, err := item.ReadJSON(bytes.NewReader(items))
	if err != nil {
		return nil, fmt.Errorf("data: items: %w", err)
	}
	return list, nil
}
