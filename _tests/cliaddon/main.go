// A test program for the cli package.
package main

import (
	"fmt"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/navigation"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
)

func main() {
	app := cli.NewApp(cli.NewStdTTY())
	navigation.Start(app, navigation.Config{
		Binding: el.MapHandler{
			term.K('x'): func() { app.CommitCode("") },
		},
	})

	code, err := app.ReadCode()
	fmt.Println("code:", code)
	if err != nil {
		fmt.Println("err", err)
	}
}
