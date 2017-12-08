// Elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"os"

	"github.com/elves/elvish/program"
)

func main() {
	os.Exit(program.Main(os.Args))
}
