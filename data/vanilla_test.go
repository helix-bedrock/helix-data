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
