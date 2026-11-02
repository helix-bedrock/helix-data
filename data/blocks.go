package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed 1.26.60/bds/climbable.json
var climbable []byte

//go:embed 1.26.60/bds/block_materials.json
var blockMaterials []byte

// MaterialWater is BDS's MaterialType of water.
const MaterialWater = 5

// Climbable returns the block types BDS lets mobs and players climb
// (ladders, vines). Each call decodes a fresh copy.
func Climbable() ([]string, error) {
	var doc struct {
		Climbable []string `json:"climbable"`
	}
	if err := json.Unmarshal(climbable, &doc); err != nil {
		return nil, fmt.Errorf("data: climbable: %w", err)
	}
	return doc.Climbable, nil
}

// BlockMaterials returns each block type's BDS MaterialType (0 air, 5
// water, 6 lava...), keyed by block name. Each call decodes a fresh copy.
func BlockMaterials() (map[string]uint8, error) {
	var doc struct {
		Blocks map[string]uint8 `json:"blocks"`
	}
	if err := json.Unmarshal(blockMaterials, &doc); err != nil {
		return nil, fmt.Errorf("data: block materials: %w", err)
	}
	return doc.Blocks, nil
}
