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
	var codec md.FmtCodec
	md.Render(string(text), &codec)
	fmt.Print(codec.String())
}
