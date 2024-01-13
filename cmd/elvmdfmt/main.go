// Command elvmdfmt reformats Markdown sources.
//
// This command is used to reformat all Markdown files in this repo; see the
// [contributor's manual] on how to use it.
//
// For general information about the Markdown implementation used by this
// command, see [src.elv.sh/pkg/md].
//
// [contributor's manual]: https://github.com/elves/elvish/blob/master/CONTRIBUTING.md#formatting
package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"os"

	"src.elv.sh/pkg/diff"
	"src.elv.sh/pkg/md"
)

var (
	overwrite = flag.Bool("w", false, "write result to source file (requires -fmt)")
	showDiff  = flag.Bool("d", false, "show diff")
	width     = flag.Int("width", 0, "if > 0, reflow content to width")
)

func main() {
	md.UnescapeHTML = html.UnescapeString
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		text, err := io.ReadAll(os.Stdin)
		handleReadError("stdin", err)
		result, unsupported := format(string(text))
		fmt.Print(result)
		handleUnsupported("stdin", unsupported)
		return
	}
	for _, file := range files {
		textBytes, err := os.ReadFile(file)
		handleReadError(file, err)
		text := string(textBytes)
		result, unsupported := format(text)
		handleUnsupported(file, unsupported)
		if *overwrite {
			err := os.WriteFile(file, []byte(result), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", file, err)
				os.Exit(2)
			}
		} else if !*showDiff {
			fmt.Print(result)
		}
		if *showDiff {
			os.Stdout.Write(diff.Diff(file+".orig", text, file, result))
		}
	}
}

func format(original string) (string, *md.FmtUnsupported) {
	codec := &md.FmtCodec{Width: *width}
	formatted := md.RenderString(original, codec)
	return formatted, codec.Unsupported()
}

func handleReadError(name string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", name, err)
		os.Exit(2)
	}
}

func handleUnsupported(name string, u *md.FmtUnsupported) {
	if u == nil {
		return
	}
	if u.NestedEmphasisOrStrongEmphasis {
		fmt.Fprintln(os.Stderr, name, "contains nested emphasis or strong emphasis")
	}
	if u.ConsecutiveEmphasisOrStrongEmphasis {
		fmt.Fprintln(os.Stderr, name, "contains consecutive emphasis or strong emphasis")
	}
	os.Exit(2)
}
