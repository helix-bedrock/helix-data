// Package item reads the item types the bds-dump add-on printed to the BDS
// console, and writes and reads them as a data file.
package item

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Item is one item type as BDS reports it.
type Item struct {
	// Name is the item's identifier, such as "minecraft:stone".
	Name string `json:"name"`
	// MaxStack is how many fit in one stack (ItemStack.maxAmount).
	MaxStack int `json:"maxStack"`
}

// logPrefix starts every item line the add-on prints.
const logPrefix = "HELIXDUMP-ITEM "

// FromLog reads the HELIXDUMP-ITEM lines of a BDS console log. Other lines
// are ignored. The result is sorted by name.
func FromLog(r io.Reader) ([]Item, error) {
	var items []Item
	seen := map[string]bool{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for line := 1; sc.Scan(); line++ {
		_, rest, ok := strings.Cut(sc.Text(), logPrefix)
		if !ok {
			continue
		}
		var it Item
		if err := json.Unmarshal([]byte(rest), &it); err != nil {
			return nil, fmt.Errorf("item: log line %d: %w", line, err)
		}
		if it.Name == "" || it.MaxStack < 1 || it.MaxStack > 255 {
			return nil, fmt.Errorf("item: log line %d: bad item %+v", line, it)
		}
		if seen[it.Name] {
			return nil, fmt.Errorf("item: log line %d: %s listed twice", line, it.Name)
		}
		seen[it.Name] = true
		items = append(items, it)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("item: read log: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("item: no %q lines in the log", strings.TrimSpace(logPrefix))
	}
	slices.SortFunc(items, func(a, b Item) int { return strings.Compare(a.Name, b.Name) })
	return items, nil
}

// WriteJSON writes items as an indented JSON array.
func WriteJSON(w io.Writer, items []Item) error {
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("item: encode: %w", err)
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

// ReadJSON reads items written by WriteJSON.
func ReadJSON(r io.Reader) ([]Item, error) {
	var items []Item
	if err := json.NewDecoder(r).Decode(&items); err != nil {
		return nil, fmt.Errorf("item: decode: %w", err)
	}
	return items, nil
}
