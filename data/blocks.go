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

//go:embed 1.26.60/bds/placement.json
var placement []byte

// PlacementTrait is one of BDS's placement traits on a block: Kind
// "cardinal" or "facing" (the direction from the player's yaw, turned by
// Offset degrees), "vertical_half" or "block_face" (from the clicked
// face and position).
type PlacementTrait struct {
	Kind   string  `json:"kind"`
	Offset float32 `json:"offset"`
}

// Placement is how BDS picks a block's state when a player places it: the
// rule its own code follows ("stairs", "pillar", "torch"...; "pending"
// when the rule needs the world and is not ported) and the traits it
// applies.
type Placement struct {
	Rule   string           `json:"rule"`
	Traits []PlacementTrait `json:"traits"`
}

// PlacementData returns the placement of every block that is not simply
// placed in its item's state, keyed by block name. Each call decodes a
// fresh copy.
func PlacementData() (map[string]Placement, error) {
	var doc struct {
		Blocks map[string]Placement `json:"blocks"`
	}
	if err := json.Unmarshal(placement, &doc); err != nil {
		return nil, fmt.Errorf("data: placement: %w", err)
	}
	return doc.Blocks, nil
}
