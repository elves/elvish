package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"src.elv.sh/pkg/md"
)

var (
	trace = flag.Bool("trace", false, "write internal parsing results")
)

func main() {
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		text, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read stdin:", err)
			os.Exit(2)
		}
		render(string(text))
		return
	}
	for _, file := range files {
		text, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", file, err)
			os.Exit(2)
		}
		render(string(text))
	}
}

func render(markdown string) {
	fmt.Print(md.RenderString(markdown, &md.HTMLCodec{}))
	if *trace {
		fmt.Print(md.RenderString(markdown, &md.TraceCodec{}))
	}
}
