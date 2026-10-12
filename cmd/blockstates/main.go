// Command blockstates reads the block states the bds-dump add-on made BDS
// save, writes them as data files, and compares them with another palette.
//
//	blockstates convert -world <bds>/worlds/helixdump -out data/1.26.52
//	blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
//	blockstates items -log <bds console log> -out data/1.26.52
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/helix-bedrock/helix-data/blockstate"
	"github.com/helix-bedrock/helix-data/item"
	"github.com/helix-bedrock/world/mcdb/leveldat"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "blockstates:", err)
		os.Exit(1) //nolint:forbidigo // a command reports failure through its exit status
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: blockstates convert|compare|experiments|items [flags]")
	}
	switch args[0] {
	case "convert":
		return convert(args[1:], out)
	case "compare":
		return compare(args[1:], out)
	case "experiments":
		return experiments(args[1:], out)
	case "items":
		return items(args[1:], out)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func convert(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	world := fs.String("world", "", "BDS world folder the dump ran in (required)")
	dir := fs.String("out", "", "output directory (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *world == "" || *dir == "" {
		return errors.New("convert needs -world and -out")
	}
	d, err := blockstate.FromWorld(filepath.Join(*world, "db"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*dir, 0o750); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(*dir, "block_states.nbt"))
	if err != nil {
		return err
	}
	if err := blockstate.WriteNBT(f, d.States); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	def, err := d.DefaultsJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*dir, "block_defaults.json"), def, 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%d blocks, %d states\n", len(d.Defaults), len(d.States))
	return err
}

// items writes items.json from the HELIXDUMP-ITEM lines of a BDS console
// log saved while the dump ran.
func items(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("items", flag.ContinueOnError)
	logPath := fs.String("log", "", "BDS console output saved while the dump ran (required)")
	dir := fs.String("out", "", "output directory (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *logPath == "" || *dir == "" {
		return errors.New("items needs -log and -out")
	}
	f, err := os.Open(*logPath)
	if err != nil {
		return err
	}
	list, err := item.FromLog(f)
	_ = f.Close()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*dir, 0o750); err != nil {
		return err
	}
	o, err := os.Create(filepath.Join(*dir, "items.json"))
	if err != nil {
		return err
	}
	if err := item.WriteJSON(o, list); err != nil {
		_ = o.Close()
		return err
	}
	if err := o.Close(); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%d items\n", len(list))
	return err
}

func compare(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	world := fs.String("world", "", "BDS world folder the dump ran in (required)")
	palette := fs.String("palette", "", "gzipped big-endian block_palette.nbt, as in CloudburstMC/Data (required)")
	limit := fs.Int("limit", 50, "differences to print per side")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *world == "" || *palette == "" {
		return errors.New("compare needs -world and -palette")
	}
	d, err := blockstate.FromWorld(filepath.Join(*world, "db"))
	if err != nil {
		return err
	}
	f, err := os.Open(*palette)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	p, err := blockstate.ReadPalette(f)
	if err != nil {
		return err
	}
	diff := blockstate.Compare(d.Plain(), p)
	// Write errors stick in the bufio.Writer and are returned by Flush.
	w := bufio.NewWriter(out)
	_, _ = fmt.Fprintf(w, "same: %d, only in dump: %d, only in palette: %d\n", diff.Same, len(diff.OnlyA), len(diff.OnlyB))
	sections := []struct {
		title string
		list  []string
	}{{"only in dump", diff.OnlyA}, {"only in palette", diff.OnlyB}}
	for _, sec := range sections {
		if len(sec.list) == 0 {
			continue
		}
		_, _ = fmt.Fprintf(w, "\n%s:\n", sec.title)
		for i, k := range sec.list {
			if i == *limit {
				_, _ = fmt.Fprintf(w, "  ... %d more\n", len(sec.list)-i)
				break
			}
			_, _ = fmt.Fprintf(w, "  %s\n", k)
		}
	}
	return w.Flush()
}

// experiments turns experiment toggles on in a world's level.dat, so a dump
// run in that world includes experimental blocks.
func experiments(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("experiments", flag.ContinueOnError)
	world := fs.String("world", "", "BDS world folder (required)")
	enable := fs.String("enable", "", "comma-separated experiment toggle names (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *world == "" || *enable == "" {
		return errors.New("experiments needs -world and -enable")
	}
	path := filepath.Join(*world, "level.dat")
	ld, err := leveldat.ReadFile(path)
	if err != nil {
		return err
	}
	var d leveldat.Data
	if err := ld.Unmarshal(&d); err != nil {
		return err
	}
	if d.Experiments == nil {
		d.Experiments = map[string]any{}
	}
	for _, name := range strings.Split(*enable, ",") {
		d.Experiments[strings.TrimSpace(name)] = uint8(1)
	}
	d.Experiments["experiments_ever_used"] = uint8(1)
	d.Experiments["saved_with_toggled_experiments"] = uint8(1)
	d.ExperimentalGamePlay = true
	if err := ld.Marshal(d); err != nil {
		return err
	}
	if err := ld.WriteFile(path); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "enabled %s in %s\n", *enable, path)
	return err
}
