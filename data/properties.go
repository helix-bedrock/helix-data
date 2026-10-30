package data

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed 1.26.60/bds/block_properties.json
var blockProperties []byte

// BlockProperties holds the numbers BDS keeps for every state of one block.
type BlockProperties struct {
	// LightEmission is the light level the block gives off, 0 to 15.
	LightEmission Values[uint8] `json:"light_emission"`
	// LightDampening is how much light the block takes away, 0 to 15.
	LightDampening Values[uint8] `json:"light_dampening"`
	// FlameOdds is how readily fire spreads onto the block.
	FlameOdds Values[uint16] `json:"flame_odds"`
	// BurnOdds is how readily fire destroys the block.
	BurnOdds Values[uint16] `json:"burn_odds"`
	// ExplosionResistance is the block's resistance to explosions.
	ExplosionResistance Values[float32] `json:"explosion_resistance"`
	// Friction is the block's friction; higher is more slippery.
	Friction Values[float32] `json:"friction"`
	// DestroyTime is the block's hardness; -1 means unbreakable.
	DestroyTime Values[float32] `json:"destroy_time"`
}

// Values is one property of a block: a single value shared by every state,
// or one value per state in the order of the block's states in
// [BlockStates].
type Values[T any] []T

// At returns the value of the block's i-th state, in [BlockStates] order.
// It returns the zero value if i is out of range.
func (v Values[T]) At(i int) T {
	if len(v) == 1 {
		return v[0]
	}
	if i >= 0 && i < len(v) {
		return v[i]
	}
	var zero T
	return zero
}

// UnmarshalJSON reads a single value or an array of values.
func (v *Values[T]) UnmarshalJSON(b []byte) error {
	if b = bytes.TrimSpace(b); len(b) > 0 && b[0] == '[' {
		var list []T
		if err := json.Unmarshal(b, &list); err != nil {
			return err
		}
		*v = list
		return nil
	}
	var one T
	if err := json.Unmarshal(b, &one); err != nil {
		return err
	}
	*v = Values[T]{one}
	return nil
}

// AllBlockProperties returns the properties of every block of [Version],
// keyed by block name, as BDS holds them at runtime. Each call decodes a
// fresh copy.
func AllBlockProperties() (map[string]BlockProperties, error) {
	var doc struct {
		Blocks map[string]BlockProperties `json:"blocks"`
	}
	if err := json.Unmarshal(blockProperties, &doc); err != nil {
		return nil, fmt.Errorf("data: block properties: %w", err)
	}
	return doc.Blocks, nil
}

//go:embed 1.26.60/block_defaults.json
var blockDefaults []byte

// BlockDefaults returns each block's default state of [Version]: block
// name to property values. JSON has one number type, so numbers (including
// the bytes boolean properties use) are float64. Each call decodes a fresh
// copy.
func BlockDefaults() (map[string]map[string]any, error) {
	var m map[string]map[string]any
	if err := json.Unmarshal(blockDefaults, &m); err != nil {
		return nil, fmt.Errorf("data: block defaults: %w", err)
	}
	return m, nil
}
