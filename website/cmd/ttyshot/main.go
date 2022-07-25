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
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 2 {
		return errors.New("Usage: ttyshot spec")
	}
	specPath := args[1]
	if !strings.HasSuffix(specPath, ".spec") {
		return fmt.Errorf("expected extension \".spec\", found %q", path.Ext(specPath))
	}
	basePath := specPath[:len(specPath)-len(".spec")]
	htmlPath := basePath + ".html"
	rawPath := basePath + ".raw"

	content, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	script, err := parseSpec(content)
	if err != nil {
		return err
	}

	outFile, err := os.OpenFile(htmlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	rawFile, err := os.OpenFile(rawPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	homePath, dbPath, cleanup, err := initEnv()
	if err != nil {
		return fmt.Errorf("set up environment: %w", err)
	}
	defer cleanup()
	return createTtyshot(homePath, dbPath, script, outFile, rawFile)
}
