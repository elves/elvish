// Command mdrun can be used to test the md package. Run it with "go run".
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"src.elv.sh/pkg/md"
)

var (
	codec = flag.String("codec", "html", "codec to use; one of html, trace, fmt, tty")
	width = flag.Int("width", 0, "text width; relevant with fmt or tty")
)

func main() {
	flag.Parse()

	c := getCodec(*codec)
	bs, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read stdin:", err)
		os.Exit(2)
	}
	fmt.Print(md.RenderString(string(bs), c))
}

func getCodec(s string) md.StringerCodec {
	switch *codec {
	case "html":
		return &md.HTMLCodec{}
	case "trace":
		return &md.TraceCodec{}
	case "fmt":
		return &md.FmtCodec{Width: *width}
	case "tty":
		return &md.TTYCodec{Width: *width}
	default:
		fmt.Println("unknown codec:", s)
		os.Exit(2)
		return nil
	}
}
