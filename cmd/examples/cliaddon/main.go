// A test program for the cli package.
package main

import (
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/mode/navigation"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

func main() {
	app := cli.NewApp(cli.AppSpec{})
	navigation.Start(app, navigation.Config{
		Binding: tk.MapHandler{
			term.K('x'): func() { app.CommitCode() },
		},
	})

	code, err := app.ReadCode()
	fmt.Println("code:", code)
	if err != nil {
		fmt.Println("err", err)
	}
}
