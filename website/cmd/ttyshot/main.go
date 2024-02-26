//go:build unix

// Command ttyshot generates a ttyshot HTML image from a ttyshot specification.
//
// See documentation in http://src.elv.sh/website#ttyshots.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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
	verboseFlag := fs.Bool("v", false, "enable verbose logging")
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

	if !*verboseFlag {
		log.SetOutput(io.Discard)
	}

	specPath := fs.Args()[0]
	content, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	script, err := parseScript(specPath, content)
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

	out, err := createTtyshot(homePath, script, *saveRawFlag)
	if err != nil {
		return err
	}

	outPath := *outFlag
	if outPath == "" {
		outPath = specPath + ".html"
	}
	return os.WriteFile(outPath, out, 0o644)
}
