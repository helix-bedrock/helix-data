package data

import (
	"encoding/json"
	"io/fs"
	"testing"
)

func TestBlockStates(t *testing.T) {
	states, err := BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	// The count the README records for 1.26.60.
	if len(states) != 23960 {
		t.Fatalf("got %d states, want 23960", len(states))
	}
	found := false
	for _, s := range states {
		if s.Name == "minecraft:air" {
			found = true
		}
	}
	if !found {
		t.Fatal("no minecraft:air")
	}
}

func TestItems(t *testing.T) {
	items, err := Items()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1623 {
		t.Fatalf("got %d items, want 1623", len(items))
	}
	want := map[string]int{"minecraft:stone": 64, "minecraft:ender_pearl": 16, "minecraft:diamond_sword": 1}
	for _, it := range items {
		if n, ok := want[it.Name]; ok && it.MaxStack != n {
			t.Errorf("%s: max stack %d, want %d", it.Name, it.MaxStack, n)
		}
	}
}

func TestEntities(t *testing.T) {
	ents, err := Entities()
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 129 {
		t.Fatalf("got %d entities, want 129", len(ents))
	}
	for _, id := range []string{"minecraft:zombie", "minecraft:cow", "minecraft:player"} {
		if _, ok := ents[id]; !ok {
			t.Errorf("no %s", id)
		}
	}
	loot, err := EntityLoot()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loot["loot_tables/entities/zombie.json"]; !ok {
		t.Error("no zombie loot table")
	}
}

func TestSpawnRules(t *testing.T) {
	rules, err := SpawnRules()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"minecraft:zombie", "minecraft:cow"} {
		if _, ok := rules[id]; !ok {
			t.Errorf("no spawn rule for %s", id)
		}
	}
}

func TestWorldgen(t *testing.T) {
	wg, err := Worldgen()
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"noise_settings/overworld", "density_function/overworld/continents", "noise/continentalness"} {
		if _, ok := wg[k]; !ok {
			t.Errorf("no %s", k)
		}
	}
	if len(BiomeTree()) < 1000 {
		t.Fatal("biome tree is empty")
	}
}

func TestFiles(t *testing.T) {
	f := Files()
	for _, p := range []string{"vanilla/biomes/plains.biome.json", "vanilla/trading/librarian_trades.json", "bds/villages.json"} {
		b, err := f.ReadFile(p)
		if err != nil || len(b) == 0 {
			t.Errorf("%s: %d bytes, %v", p, len(b), err)
		}
	}
	n := 0
	if err := fs.WalkDir(f, "vanilla", func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			n++
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	// The count cmd/vanilla printed for 1.26.60.6.
	if n != 4270 {
		t.Errorf("got %d vanilla files, want 4270", n)
	}
	if _, err := f.Open("../vanilla.go"); err == nil {
		t.Error("opened a path outside the tree")
	}
}

func TestFilesSpawnRules(t *testing.T) {
	b, err := fs.ReadFile(Files(), "spawn_rules.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty spawn_rules.json")
	}
}

func TestFilesBlockMaterials(t *testing.T) {
	b, err := fs.ReadFile(Files(), "bds/block_materials.json")
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Materials []struct {
			Type int `json:"type"`
		} `json:"materials"`
		Blocks map[string]int `json:"blocks"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Blocks) < 1400 || len(m.Materials) == 0 {
		t.Fatalf("%d blocks, %d materials", len(m.Blocks), len(m.Materials))
	}
	if m.Blocks["minecraft:water"] != 5 || m.Blocks["minecraft:air"] != 0 {
		t.Fatalf("water %d, air %d", m.Blocks["minecraft:water"], m.Blocks["minecraft:air"])
	}
}

// TestBlockProperties checks that every block state has properties and
// that per-state arrays match the block's state count.
func TestBlockProperties(t *testing.T) {
	props, err := AllBlockProperties()
	if err != nil {
		t.Fatal(err)
	}
	states, err := BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	count := map[string]int{}
	for _, s := range states {
		count[s.Name]++
	}
	if len(props) != len(count) {
		t.Fatalf("%d blocks with properties, %d blocks with states", len(props), len(count))
	}
	for name, n := range count {
		p, ok := props[name]
		if !ok {
			t.Fatalf("%s has no properties", name)
		}
		if l := len(p.LightEmission); l != 1 && l != n {
			t.Fatalf("%s: %d light emission values for %d states", name, l, n)
		}
	}
	if got := props["minecraft:stone"].LightDampening.At(0); got != 15 {
		t.Fatalf("stone light dampening %d, want 15", got)
	}
}

// TestItemRegistry checks BDS's item list against items.json.
func TestItemRegistry(t *testing.T) {
	reg, err := ItemRegistry()
	if err != nil {
		t.Fatal(err)
	}
	list, err := Items()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[int16]string{}
	names := map[string]bool{}
	for _, it := range reg {
		if other, dup := ids[it.RuntimeID]; dup {
			t.Fatalf("%s and %s share ID %d", it.Name, other, it.RuntimeID)
		}
		ids[it.RuntimeID] = it.Name
		names[it.Name] = true
	}
	missing := 0
	for _, it := range list {
		if !names[it.Name] {
			missing++
		}
	}
	t.Logf("%d items in the registry, %d in items.json, %d of those not in the registry", len(reg), len(list), missing)
}
