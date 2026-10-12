// Command blockstates turns a BDS console log from the bds-dump add-on into
// data files, and compares the result with another palette.
//
//	blockstates convert -log bds.log -version 18168865 -out data/1.26.52
//	blockstates compare -dump bds.log -palette block_palette.nbt
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func readDump(path string) (*blockstate.Dump, error) {
	f, err := os.Open(path) //nolint:gosec // path comes from the command line
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return blockstate.ReadDump(f)
}

func convert(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	logPath := fs.String("log", "", "BDS console log containing the dump")
	version := fs.Int("version", 0, "block version number to store in every state (required)")
	dir := fs.String("out", "", "output directory (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *logPath == "" || *dir == "" || *version <= 0 {
		return errors.New("convert needs -log, -version and -out")
	}
	d, err := readDump(*logPath)
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
	if err := blockstate.WriteNBT(f, d.States, int32(*version)); err != nil { //nolint:gosec // checked > 0; versions fit in int32
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
	_, err = fmt.Fprintf(out, "%d blocks, %d states, %d skipped combinations\n", len(d.Defaults), len(d.States), len(d.Skipped))
	return err
}

func compare(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	logPath := fs.String("dump", "", "BDS console log containing the dump")
	palette := fs.String("palette", "", "gzipped big-endian block_palette.nbt (CloudburstMC/Data)")
	limit := fs.Int("limit", 50, "differences to print per side")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *logPath == "" || *palette == "" {
		return errors.New("compare needs -dump and -palette")
	}
	d, err := readDump(*logPath)
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
	diff := blockstate.Compare(d.States, p)
	// Write errors stick in the bufio.Writer and are returned by Flush.
	w := bufio.NewWriter(out)
	_, _ = fmt.Fprintf(w, "same: %d, only in dump: %d, only in palette: %d, skipped by BDS: %d\n",
		diff.Same, len(diff.OnlyA), len(diff.OnlyB), len(d.Skipped))
	sections := []struct {
		title string
		list  []string
	}{{"only in dump", diff.OnlyA}, {"only in palette", diff.OnlyB}, {"skipped by BDS", d.Skipped}}
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
