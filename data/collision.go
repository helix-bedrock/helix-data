package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed 1.26.60/bds/collision.json
var collision []byte

// Box is a collision box inside the block, min x, y, z then max x, y, z.
type Box [6]float32

// Collision is BDS's block collision data: the boxes of the blocks whose
// shape BDS stores as data, and the blocks whose shape BDS computes in code.
type Collision struct {
	// Blocks are the blocks whose shape is data, keyed by name.
	Blocks map[string]CollisionBlock `json:"blocks"`
	// Computed lists the blocks whose shape BDS computes in code, with the
	// methods their block type replaces ("collision", "shapes", "visual").
	// Their shapes are not in this file.
	Computed map[string][]string `json:"computed"`
}

// CollisionBlock is one block's collision: Boxes for every state, or, when
// its states differ, a list per state. No boxes means no collision.
type CollisionBlock struct {
	Boxes  []Box            `json:"boxes"`
	States []CollisionState `json:"states"`
}

// CollisionState is the boxes of the state with the given properties.
type CollisionState struct {
	Properties map[string]any
	Boxes      []Box
}

// UnmarshalJSON reads a state written as {"states": {name: [type, value]},
// "boxes": [...]}, keeping only the values.
func (s *CollisionState) UnmarshalJSON(b []byte) error {
	var raw struct {
		States map[string][2]json.RawMessage `json:"states"`
		Boxes  []Box                         `json:"boxes"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	s.Boxes, s.Properties = raw.Boxes, make(map[string]any, len(raw.States))
	for k, tv := range raw.States {
		var kind string
		if err := json.Unmarshal(tv[0], &kind); err != nil {
			return fmt.Errorf("state %q: %w", k, err)
		}
		if kind == "string" {
			var v string
			if err := json.Unmarshal(tv[1], &v); err != nil {
				return fmt.Errorf("state %q: %w", k, err)
			}
			s.Properties[k] = v
			continue
		}
		var v int32
		if err := json.Unmarshal(tv[1], &v); err != nil {
			return fmt.Errorf("state %q: %w", k, err)
		}
		s.Properties[k] = v
	}
	return nil
}

// CollisionData returns the collision table. Each call decodes a fresh copy.
func CollisionData() (*Collision, error) {
	var c Collision
	if err := json.Unmarshal(collision, &c); err != nil {
		return nil, fmt.Errorf("data: collision: %w", err)
	}
	return &c, nil
}
