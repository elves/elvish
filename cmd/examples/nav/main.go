// Command nav runs the navigation mode of the line editor.
package main

import (
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

func main() {
	app := cli.NewApp(cli.AppSpec{})
	w, _ := modes.NewNavigation(app, modes.NavigationSpec{
		Bindings: tk.MapBindings{
			term.K('x'): func(tk.Widget) { app.CommitCode() },
		},
	})
	app.PushAddon(w)

	code, err := app.ReadCode()
	fmt.Println("code:", code)
	if err != nil {
		fmt.Println("err", err)
	}
}
