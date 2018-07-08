package main

import (
	"fmt"
	"os"

	"github.com/elves/elvish/newedit"
)

func main() {
	ed := newedit.NewEditor(os.Stdin, os.Stdout)
	buf, err := ed.ReadLine()
	fmt.Println("buffer:", buf, "error:", err)
}
