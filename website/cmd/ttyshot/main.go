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
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Expected one argument, got %d\n", len(os.Args)-1)
		os.Exit(1)
	}
	specPath := os.Args[1]
	if !strings.HasSuffix(specPath, ".spec") {
		fmt.Fprintf(os.Stderr, "Expected extension \".spec\", found %q\n", path.Ext(specPath))
		os.Exit(2)
	}
	basePath := specPath[:len(specPath)-len(".spec")]
	htmlPath := basePath + ".html"
	rawPath := basePath + ".raw"

	content, err := os.ReadFile(specPath)
	if err != nil {
		log.Fatal(err)
	}

	script, err := parseSpec(content)
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.OpenFile(htmlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	rawFile, err := os.OpenFile(rawPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	homePath, dbPath, cleanup := initEnv()
	defer cleanup()
	if err := createTtyshot(homePath, dbPath, script, outFile, rawFile); err != nil {
		log.Fatal(err)
	}
}
