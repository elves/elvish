package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"src.elv.sh/pkg/md"
)

func main() {
	text, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(md.RenderString(string(text), &md.TraceCodec{}))
}
