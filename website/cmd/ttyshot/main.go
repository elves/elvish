// Generate a ttyshot HTML image from a ttyshot specification.
//
// Usage: ./ttyshot website/ttyshot/*.spec
//
// You can recreate all the ttyshots by running the following from the project top-level directory:
//
//   make ttyshot
//   for f [website/ttyshot/**.spec] { put $f; ./ttyshot $f >/dev/tty 2>&1 }
//
// This assumes working `elvish` and `tmux` programs in $E:PATH.
//
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ttyshot", flag.ExitOnError)
	outFlag := fs.String("o", "", "output file (defaults to spec-path + .html)")
	saveRawFlag := fs.String("save-raw", "", "if non-empty, save output of capture-pane to this file")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: ttyshot [flags] spec-path")
		fs.PrintDefaults()
	}

	fs.Parse(args)
	if len(fs.Args()) != 1 {
		fs.Usage()
		os.Exit(1)
	}

	specPath := fs.Args()[0]
	content, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	spec, err := parseSpec(content)
	if err != nil {
		return err
	}

	homePath, err := setupHome()
	if err != nil {
		return fmt.Errorf("set up temporary home: %w", err)
	}
	defer func() {
		err := os.RemoveAll(homePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: unable to remove temp HOME:", err)
		}
	}()

	out, err := createTtyshot(homePath, spec, *saveRawFlag)
	if err != nil {
		return err
	}

	outPath := *outFlag
	if outPath == "" {
		outPath = specPath + ".html"
	}
	return os.WriteFile(outPath, out, 0o644)
}
