package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"src.elv.sh/pkg/md"
)

func main() {
	var (
		format = flag.Bool("fmt", false, "format Markdown")
		trace  = flag.Bool("trace", false, "trace internal output by parser")
	)
	flag.Parse()
	if *format && *trace {
		fmt.Fprintln(os.Stderr, "-fmt and -trace are mutually exclusive")
		os.Exit(1)
	}

	text, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	var codec md.StringerCodec
	switch {
	case *format:
		codec = &md.FmtCodec{}
	case *trace:
		codec = &md.TraceCodec{}
	default:
		codec = &md.HTMLCodec{}
	}
	fmt.Print(md.RenderString(string(text), codec))
}
