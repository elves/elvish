// A test program for the cli package.
package main

import (
	"fmt"
	"io"

	"github.com/elves/elvish/cli/clicore"
)

func main() {
	app := clicore.NewAppFromStdIO()
	for {
		code, err := app.ReadCode()
		if err != nil {
			if err != io.EOF {
				fmt.Println("error:", err)
			}
			break
		}
		fmt.Println("got:", code)
		if code == "exit" {
			fmt.Println("bye")
			break
		}
	}
}
