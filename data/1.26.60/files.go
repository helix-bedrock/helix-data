// Package files embeds the vanilla behaviour pack files and BDS tables of
// Minecraft 1.26.60 (BDS 1.26.60.6). It uses only the standard library, so
// programs that need just these files, such as the world generator, do not
// pull in the rest of helix-data's dependencies.
//
// This file is the only hand-written one in the folder; cmd/vanilla
// regenerates vanilla/ and leaves it alone.
package files

import (
	"embed"
	"io/fs"
)

//go:embed vanilla experiments bds spawn_rules.json
var tree embed.FS

// FS is the read-only file tree: "vanilla/..." holds the vanilla behaviour
// pack folders (biomes, features, feature_rules, loot_tables, recipes,
// spawn_groups, structures, trading, worldgen) as BDS stacks them,
// "experiments/<name>/..." the same folders of each experimental pack on
// its own (such as "y_2026_drop_4"; a world with the experiment on reads
// them over vanilla/), "bds/..." the tables extracted from BDS's code, and
// spawn_rules.json every entity's spawn rules file keyed by identifier.
// Files whose name starts with "_" or "." (such as _provenance.tsv) are not
// embedded.
type FS struct{}

// Files returns the file tree.
func Files() FS { return FS{} }

// Open implements fs.FS.
func (FS) Open(name string) (fs.File, error) { return tree.Open(name) }

// ReadFile implements fs.ReadFileFS. The bytes are a fresh copy.
func (FS) ReadFile(name string) ([]byte, error) { return tree.ReadFile(name) }

// ReadDir implements fs.ReadDirFS.
func (FS) ReadDir(name string) ([]fs.DirEntry, error) { return tree.ReadDir(name) }
