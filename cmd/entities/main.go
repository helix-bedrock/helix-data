// Command entities writes entities.json, entity_loot.json and
// spawn_rules.json from a BDS install: every vanilla entity definition,
// entity loot table and spawn rule, with the
// vanilla behaviour packs layered the way BDS layers them.
//
//	go run ./cmd/entities -bds <bds folder> -out data/<version>
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
	"strconv"
	"strings"

	"github.com/helix-bedrock/helix-data/brarchive"
)

func main() {
	bds := flag.String("bds", "", "BDS folder")
	out := flag.String("out", "", "output folder")
	flag.Parse()
	if *bds == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*bds, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(bds, out string) error {
	packs, err := vanillaPacks(filepath.Join(bds, "behavior_packs"))
	if err != nil {
		return err
	}
	entities, loot, spawns := map[string]json.RawMessage{}, map[string]json.RawMessage{}, map[string]json.RawMessage{}
	for _, p := range packs {
		if err := readArchive(p, "entities.brarchive", "minecraft:entity", entities); err != nil {
			return err
		}
		if err := readArchive(p, "spawn_rules.brarchive", "minecraft:spawn_rules", spawns); err != nil {
			return err
		}
		if err := readLoot(p, loot); err != nil {
			return err
		}
	}
	if err := write(filepath.Join(out, "entities.json"), entities); err != nil {
		return err
	}
	if err := write(filepath.Join(out, "entity_loot.json"), loot); err != nil {
		return err
	}
	if err := write(filepath.Join(out, "spawn_rules.json"), spawns); err != nil {
		return err
	}
	fmt.Printf("%d entities, %d loot tables, %d spawn rules from %d packs\n", len(entities), len(loot), len(spawns), len(packs))
	return nil
}

// vanillaPacks returns the vanilla packs in the order BDS stacks them:
// "vanilla" first, then "vanilla_<version>" from oldest to newest, so a
// later pack overrides an earlier one.
func vanillaPacks(dir string) ([]string, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read behaviour packs: %w", err)
	}
	type pack struct {
		path string
		ver  []int
	}
	var packs []pack
	for _, de := range des {
		name := de.Name()
		switch {
		case name == "vanilla":
			packs = append(packs, pack{path: filepath.Join(dir, name)})
		case strings.HasPrefix(name, "vanilla_"):
			var ver []int
			for _, s := range strings.Split(strings.TrimPrefix(name, "vanilla_"), ".") {
				n, err := strconv.Atoi(s)
				if err != nil {
					return nil, fmt.Errorf("pack %s: bad version", name)
				}
				ver = append(ver, n)
			}
			packs = append(packs, pack{path: filepath.Join(dir, name), ver: ver})
		}
	}
	slices.SortFunc(packs, func(a, b pack) int { return slices.Compare(a.ver, b.ver) })
	out := make([]string, len(packs))
	for i, p := range packs {
		out[i] = p.path
	}
	return out, nil
}

// readArchive reads the JSON files of one archive, keyed by the identifier
// in their description under root, such as "minecraft:entity".
func readArchive(pack, archive, root string, into map[string]json.RawMessage) error {
	b, err := os.ReadFile(filepath.Join(pack, "__brarchive", archive))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	files, err := brarchive.Read(b)
	if err != nil {
		return fmt.Errorf("%s: %w", pack, err)
	}
	for _, f := range files {
		var doc map[string]json.RawMessage
		if err := json.Unmarshal(f.Data, &doc); err != nil {
			return fmt.Errorf("%s: %s: %w", pack, f.Name, err)
		}
		var body struct {
			Description struct {
				Identifier string `json:"identifier"`
			} `json:"description"`
		}
		if raw, ok := doc[root]; ok {
			if err := json.Unmarshal(raw, &body); err != nil {
				return fmt.Errorf("%s: %s: %w", pack, f.Name, err)
			}
		}
		id := body.Description.Identifier
		if id == "" {
			// vanilla_1.16.100 has an npc.json with only a component and no
			// identifier, so it cannot apply to any entity.
			continue
		}
		into[id] = json.RawMessage(bytes.Clone(f.Data))
	}
	return nil
}

// readLoot reads loot_tables/entities, keyed by the path entity
// definitions use, such as "loot_tables/entities/zombie.json".
func readLoot(pack string, into map[string]json.RawMessage) error {
	dir := filepath.Join(pack, "loot_tables", "entities")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !json.Valid(b) {
			return fmt.Errorf("%s: invalid JSON", path)
		}
		rel, err := filepath.Rel(pack, path)
		if err != nil {
			return err
		}
		into[filepath.ToSlash(rel)] = json.RawMessage(b)
		return nil
	})
}

// write writes m as compact JSON, one key per line, sorted, so diffs
// between versions stay readable.
func write(path string, m map[string]json.RawMessage) error {
	keys := slices.Sorted(maps.Keys(m))
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, k := range keys {
		var c bytes.Buffer
		if err := json.Compact(&c, m[k]); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(c.Bytes())
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
