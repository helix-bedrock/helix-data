package data

import "testing"

func TestBlockStates(t *testing.T) {
	states, err := BlockStates()
	if err != nil {
		t.Fatal(err)
	}
	// The count the README records for 1.26.52.
	if len(states) != 22091 {
		t.Fatalf("got %d states, want 22091", len(states))
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
