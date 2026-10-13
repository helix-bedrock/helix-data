// Package data gives Go programs the generated game data in data/, so the
// engine can load it without a copy of this repo on disk.
package data

import (
	"bytes"
	_ "embed"
	"encoding/json"
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

//go:embed 1.26.52/entities.json
var entities []byte

// Entities returns every vanilla entity definition of [Version], keyed by
// identifier such as "minecraft:zombie". Each value is the definition file
// BDS uses, with the vanilla behaviour packs layered as BDS layers them.
// Each call decodes a fresh copy.
func Entities() (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(entities, &m); err != nil {
		return nil, fmt.Errorf("data: entities: %w", err)
	}
	return m, nil
}

//go:embed 1.26.52/entity_loot.json
var entityLoot []byte

// EntityLoot returns every entity loot table of [Version], keyed by the
// path entity definitions name them by, such as
// "loot_tables/entities/zombie.json". Each call decodes a fresh copy.
func EntityLoot() (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(entityLoot, &m); err != nil {
		return nil, fmt.Errorf("data: entity loot: %w", err)
	}
	return m, nil
}

//go:embed 1.26.52/spawn_rules.json
var spawnRules []byte

// SpawnRules returns every vanilla spawn rule of [Version], keyed by entity
// identifier such as "minecraft:zombie". Each value is the spawn rule file
// BDS uses. Each call decodes a fresh copy.
func SpawnRules() (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(spawnRules, &m); err != nil {
		return nil, fmt.Errorf("data: spawn rules: %w", err)
	}
	return m, nil
}
