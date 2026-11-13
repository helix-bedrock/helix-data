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
