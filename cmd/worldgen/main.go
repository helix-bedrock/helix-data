// Command worldgen writes worldgen.json and biome_tree.json: the world
// generation data Helix uses for terrain.
//
// BDS does not ship its world generator's data: it is compiled into the
// server. Since 1.18 Bedrock and Java Edition generate the same terrain
// shape and biomes for a seed, so Helix uses Java Edition's data, from the
// vanilla datapack inside the official Java server jar:
//
//	go run ./cmd/worldgen -datapack <extracted datapack>/data/minecraft/worldgen \
//	    -tree <multi_noise_biome_tree.json> -out data/<version>
//
// The overworld's biome parameters are not a datapack file in Java
// Edition but built by its code; -tree is that table as Java's search tree,
// extracted from Java Edition by Pumpkin's extractor.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// folders are the parts of the datapack's worldgen folder terrain needs.
var folders = []string{"density_function", "noise", "noise_settings", "material_rule", "material_condition"}

func main() {
	pack := flag.String("datapack", "", "the datapack's data/minecraft/worldgen folder")
	tree := flag.String("tree", "", "multi_noise_biome_tree.json")
	out := flag.String("out", "", "output folder")
	flag.Parse()
	if *pack == "" || *tree == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*pack, *tree, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(pack, tree, out string) error {
	files := map[string]json.RawMessage{}
	for _, f := range folders {
		root := filepath.Join(pack, f)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
				return err
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var c bytes.Buffer
			if err := json.Compact(&c, b); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			rel, err := filepath.Rel(pack, path)
			if err != nil {
				return err
			}
			// Keyed as Minecraft names them: "density_function/overworld/depth".
			files[strings.TrimSuffix(filepath.ToSlash(rel), ".json")] = c.Bytes()
			return nil
		})
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
	}
	if err := write(filepath.Join(out, "worldgen.json"), files); err != nil {
		return err
	}
	b, err := os.ReadFile(tree)
	if err != nil {
		return err
	}
	var c bytes.Buffer
	if err := json.Compact(&c, b); err != nil {
		return fmt.Errorf("%s: %w", tree, err)
	}
	if err := os.WriteFile(filepath.Join(out, "biome_tree.json"), append(c.Bytes(), '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("%d world generation files and the biome tree\n", len(files))
	return nil
}

// write writes m as JSON, one key per line, sorted.
func write(path string, m map[string]json.RawMessage) error {
	var buf bytes.Buffer
	buf.WriteString("{\n")
	keys := slices.Sorted(maps.Keys(m))
	for i, k := range keys {
		kb, err := json.Marshal(k)
		if err != nil {
			return err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(m[k])
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
