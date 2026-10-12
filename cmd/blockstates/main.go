// Command blockstates reads the block states the bds-dump add-on made BDS
// save, writes them as data files, and compares them with another palette.
//
//	blockstates convert -world <bds>/worlds/helixdump -out data/1.26.52
//	blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
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
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "blockstates:", err)
		os.Exit(1) //nolint:forbidigo // a command reports failure through its exit status
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: blockstates convert|compare [flags]")
	}
	switch args[0] {
	case "convert":
		return convert(args[1:], out)
	case "compare":
		return compare(args[1:], out)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func convert(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	world := fs.String("world", "", "BDS world folder the dump ran in (required)")
	dir := fs.String("out", "", "output directory (required)")
	fill := fs.String("fill", "", "palette (CloudburstMC/Data block_palette.nbt) to add the -fill-blocks states from")
	fillBlocks := fs.String("fill-blocks", blockstate.DefaultFillBlocks, "comma-separated blocks whose states the dump cannot produce")
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
	if *fill != "" {
		f, err := os.Open(*fill)
		if err != nil {
			return err
		}
		p, err := blockstate.ReadPalette(f)
		_ = f.Close()
		if err != nil {
			return err
		}
		n, err := d.Fill(p, strings.Split(*fillBlocks, ","))
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "added %d states from %s for %s\n", n, *fill, *fillBlocks)
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
