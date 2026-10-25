// Package pack reads BDS's vanilla behaviour packs and stacks them the way
// BDS does for a world on the server's own version.
//
// BDS ships one pack per game version ("vanilla", "vanilla_1.14", ...,
// "vanilla_1.26.52"), and each pack holds only the files that changed in
// that version. BDS layers them over its definitions/ folder, oldest first,
// and a later file replaces an earlier one at the same path. Most files sit
// in .brarchive files under __brarchive/, which mirror the folder tree:
// __brarchive/loot_tables/chests.brarchive holds loot_tables/chests/*.
package pack

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/helix-bedrock/helix-data/brarchive"
)

// Definitions is the From of files from BDS's definitions/ folder.
const Definitions = "definitions"

// File is one file of the stacked packs.
type File struct {
	Data []byte
	// From is the pack the file came from, such as "vanilla_1.26.50", or
	// [Definitions].
	From string
}

// Vanilla returns the vanilla pack folders under dir in the order BDS
// stacks them: "vanilla" first, then "vanilla_<version>" from oldest to
// newest.
func Vanilla(dir string) ([]string, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("pack: read behaviour packs: %w", err)
	}
	type entry struct {
		path string
		ver  []int
	}
	var packs []entry
	for _, de := range des {
		name := de.Name()
		switch {
		case name == "vanilla":
			packs = append(packs, entry{path: filepath.Join(dir, name)})
		case strings.HasPrefix(name, "vanilla_"):
			var ver []int
			for s := range strings.SplitSeq(strings.TrimPrefix(name, "vanilla_"), ".") {
				n, err := strconv.Atoi(s)
				if err != nil {
					return nil, fmt.Errorf("pack: %s: bad version", name)
				}
				ver = append(ver, n)
			}
			packs = append(packs, entry{path: filepath.Join(dir, name), ver: ver})
		}
	}
	slices.SortFunc(packs, func(a, b entry) int { return slices.Compare(a.ver, b.ver) })
	out := make([]string, len(packs))
	for i, p := range packs {
		out[i] = p.path
	}
	return out, nil
}

// Stack returns the files of the top-level folders dirs (such as "biomes"
// or "loot_tables") as BDS sees them, keyed by slash path such as
// "loot_tables/chests/village/village_bakery.json". bds is the BDS folder.
func Stack(bds string, dirs []string) (map[string]File, error) {
	out := map[string]File{}
	keep := func(p string) bool {
		top, _, _ := strings.Cut(p, "/")
		return slices.Contains(dirs, top)
	}
	if err := readLoose(filepath.Join(bds, Definitions), Definitions, keep, out); err != nil {
		return nil, err
	}
	packs, err := Vanilla(filepath.Join(bds, "behavior_packs"))
	if err != nil {
		return nil, err
	}
	for _, p := range packs {
		name := filepath.Base(p)
		files := map[string]File{}
		if err := readArchives(p, name, keep, files); err != nil {
			return nil, err
		}
		// The plain files go second: archives list binary structures as
		// empty entries and the real file sits next to them.
		if err := readLoose(p, name, keep, files); err != nil {
			return nil, err
		}
		for k, f := range files {
			out[k] = f
		}
	}
	return out, nil
}

// readLoose reads the plain files in the folders of root. Files directly in
// root (manifests, icons, stray JSON) belong to no folder and are skipped.
func readLoose(root, from string, keep func(string) bool, into map[string]File) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "__brarchive" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.Contains(rel, "/") || !keep(rel) {
			return nil
		}
		b, err := os.ReadFile(p) //nolint:gosec // p comes from walking the BDS folder the caller named
		if err != nil {
			return fmt.Errorf("pack: %w", err)
		}
		into[rel] = File{Data: b, From: from}
		return nil
	})
}

// readArchives reads __brarchive/<dir>.brarchive as the files of <dir>.
func readArchives(root, from string, keep func(string) bool, into map[string]File) error {
	base := filepath.Join(root, "__brarchive")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".brarchive") {
			return err
		}
		rel, err := filepath.Rel(base, p)
		if err != nil {
			return err
		}
		dir := strings.TrimSuffix(filepath.ToSlash(rel), ".brarchive")
		if !keep(dir) {
			return nil
		}
		b, err := os.ReadFile(p) //nolint:gosec // p comes from walking the BDS folder the caller named
		if err != nil {
			return fmt.Errorf("pack: %w", err)
		}
		files, err := brarchive.Read(b)
		if err != nil {
			return fmt.Errorf("pack: %s: %w", p, err)
		}
		for _, f := range files {
			if len(f.Data) == 0 {
				continue
			}
			into[path.Join(dir, f.Name)] = File{Data: f.Data, From: from}
		}
		return nil
	})
}
