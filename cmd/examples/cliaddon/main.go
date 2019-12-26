// A test program for the cli package.
package main

import (
	"fmt"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/addons/navigation"
	"github.com/elves/elvish/pkg/cli/term"
)

func main() {
	app := cli.NewApp(cli.AppSpec{})
	navigation.Start(app, navigation.Config{
		Binding: cli.MapHandler{
			term.K('x'): func() { app.CommitCode() },
		},
	})

	code, err := app.ReadCode()
	fmt.Println("code:", code)
	if err != nil {
		fmt.Println("err", err)
	}
}
