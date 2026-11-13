package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed files/bds/block_solid.json
var blockSolid []byte

// SolidBlocks returns the block types whose BDS Block solid flag is set: full opaque
// blocks such as stone, dirt and glowstone, not glass, ice, leaves, slabs or stairs. It is
// neither the material's solid flag nor the one a ladder checks; a respawn anchor's spawn
// search stands players on these. Each call decodes a fresh copy.
func SolidBlocks() ([]string, error) {
	var doc struct {
		Solid []string `json:"solid"`
	}
	if err := json.Unmarshal(blockSolid, &doc); err != nil {
		return nil, fmt.Errorf("data: block solid: %w", err)
	}
	return doc.Solid, nil
}

//go:embed files/bds/suffocating.json
var blockSuffocating []byte

// SuffocatingBlocks returns the block types BDS's breathing check calls solid: a mob or
// player whose breathing point is inside one has no air. They are the SolidBlocks less those
// whose block type lacks two more of BDS's flags (camera, jigsaw, structure block, honey
// block, the client request placeholder). Glass, ice, leaves, slabs and stairs are not among
// them. Each call decodes a fresh copy.
func SuffocatingBlocks() ([]string, error) {
	var doc struct {
		Suffocating []string `json:"suffocating"`
	}
	if err := json.Unmarshal(blockSuffocating, &doc); err != nil {
		return nil, fmt.Errorf("data: suffocating blocks: %w", err)
	}
	return doc.Suffocating, nil
}
