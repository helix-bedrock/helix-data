package data

import (
	"bytes"
	_ "embed"
)

//go:embed files/bds/commands.json
var commands []byte

// Commands is BDS's command list as it sends it to an operator: every command with its
// description, permission level, flags, aliases and overloads, each argument with its name,
// type and enum, and the enums and soft enums the arguments use. Lists the engine fills from
// its own registries (items, blocks, entity types, biomes, recipes) are left empty. Helix
// generates its vanilla command syntax from it. Each call returns a fresh copy.
func Commands() []byte { return bytes.Clone(commands) }
