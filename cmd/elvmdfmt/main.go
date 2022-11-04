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
)

func main() {
	md.UnescapeHTML = html.UnescapeString
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		text, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read stdin:", err)
			os.Exit(2)
		}
		fmt.Print(format(string(text)))
		return
	}
	for _, file := range files {
		text, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", file, err)
			os.Exit(2)
		}
		result := format(string(text))
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
			os.Stdout.Write(diff.Diff(file+".orig", text, file, []byte(result)))
		}
	}
}

func format(original string) string {
	return md.RenderString(original, &md.FmtCodec{})
}
