// Command vanilla writes <out>/vanilla from a BDS install: the vanilla
// behaviour pack folders Helix reads as plain files, with every pack
// stacked the way BDS stacks them, and _provenance.tsv naming the pack each
// file came from.
//
//	go run ./cmd/vanilla -bds <bds folder> -out data/<version>
package main

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/helix-bedrock/helix-data/pack"
)

// dirs are the pack folders Helix reads. Entities and spawn rules are in
// entities.json and spawn_rules.json instead.
var dirs = []string{
	"biomes", "feature_rules", "features", "loot_tables", "recipes",
	"spawn_groups", "structures", "trading", "worldgen",
}

// pinned are files BDS 1.26.52.3 takes from an older pack than the stack
// gives, as measured by comparing generated villages with BDS's (worldgen
// commit 8566743). vanilla_1.20.0 changes the three desert meeting points,
// but BDS builds desert villages with the base pack's.
var pinned = map[string]string{
	"structures/village/desert/town_centers/desert_meeting_point_1.nbt": "vanilla",
	"structures/village/desert/town_centers/desert_meeting_point_2.nbt": "vanilla",
	"structures/village/desert/town_centers/desert_meeting_point_3.nbt": "vanilla",
}

func main() {
	bds := flag.String("bds", "", "BDS folder")
	out := flag.String("out", "", "output folder")
	flag.Parse()
	if *bds == "" || *out == "" {
		flag.Usage()
		os.Exit(2) //nolint:forbidigo // a command reports failure through its exit status
	}
	if err := run(*bds, filepath.Join(*out, "vanilla")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1) //nolint:forbidigo // a command reports failure through its exit status
	}
	if err := runExperiments(*bds, filepath.Join(*out, "experiments")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1) //nolint:forbidigo // a command reports failure through its exit status
	}
}

// runExperiments writes each experimental pack's own files to
// <out>/<name>, where name is the pack folder without "experimental_". The
// files are not stacked: a world with the experiment on reads them over
// vanilla/.
func runExperiments(bds, out string) error {
	packs, err := pack.Experiments(filepath.Join(bds, "behavior_packs"))
	if err != nil {
		return err
	}
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	for _, p := range packs {
		files, err := pack.Read(p, dirs)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			continue
		}
		name := strings.TrimPrefix(filepath.Base(p), pack.ExperimentPrefix)
		if err := write(filepath.Join(out, name), files); err != nil {
			return err
		}
		fmt.Printf("experiment %s: %d files\n", name, len(files))
	}
	return nil
}

// write puts files under out with _provenance.tsv naming each file's pack.
func write(out string, files map[string]pack.File) error {
	var prov strings.Builder
	for _, k := range slices.Sorted(maps.Keys(files)) {
		f := files[k]
		p := filepath.Join(out, filepath.FromSlash(k))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil { //nolint:gosec // generated data folders are committed and must be readable
			return err
		}
		if err := os.WriteFile(p, f.Data, 0o644); err != nil { //nolint:gosec // generated data files are committed and must be readable
			return err
		}
		fmt.Fprintf(&prov, "%s\t%s\n", k, f.From)
	}
	return os.WriteFile(filepath.Join(out, "_provenance.tsv"), []byte(prov.String()), 0o644) //nolint:gosec // generated data files are committed and must be readable
}

func run(bds, out string) error {
	files, err := pack.Stack(bds, dirs)
	if err != nil {
		return err
	}
	for k, from := range pinned {
		b, err := os.ReadFile(filepath.Join(bds, "behavior_packs", from, filepath.FromSlash(k))) //nolint:gosec // the path is one the person passed on the command line
		if err != nil {
			return fmt.Errorf("pinned %s: %w", k, err)
		}
		files[k] = pack.File{Data: b, From: from}
	}
	// The folder is generated, so stale files from an older run must go.
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	if err := write(out, files); err != nil {
		return err
	}
	fmt.Printf("%d files\n", len(files))
	return nil
}
