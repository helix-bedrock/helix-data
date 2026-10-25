// Package files embeds the vanilla behaviour pack files and BDS tables of
// Minecraft 1.26.52 (BDS 1.26.52.3). It uses only the standard library, so
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

//go:embed vanilla bds
var tree embed.FS

// FS is the read-only file tree: "vanilla/..." holds the vanilla behaviour
// pack folders (biomes, features, feature_rules, loot_tables, recipes,
// spawn_groups, structures, trading, worldgen) as BDS stacks them, and
// "bds/..." the tables extracted from BDS's code. Files whose name starts
// with "_" or "." (such as _provenance.tsv) are not embedded.
type FS struct{}

// Files returns the file tree.
func Files() FS { return FS{} }

// Open implements fs.FS.
func (FS) Open(name string) (fs.File, error) { return tree.Open(name) }

// ReadFile implements fs.ReadFileFS. The bytes are a fresh copy.
func (FS) ReadFile(name string) ([]byte, error) { return tree.ReadFile(name) }

// ReadDir implements fs.ReadDirFS.
func (FS) ReadDir(name string) ([]fs.DirEntry, error) { return tree.ReadDir(name) }
