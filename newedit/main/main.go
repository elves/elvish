package main

import (
	"fmt"

	"github.com/elves/elvish/newedit/core"
)

func main() {
	ed := core.NewStdEditor()
	buf, err := ed.ReadCode()
	fmt.Println("buffer:", buf, "error:", err)
}
