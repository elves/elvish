// Command e3bc ("Elvish-editor-enhanced bc") is a wrapper for the bc command
// that uses Elvish's cli library for an enhanced CLI experience.
package main

import (
	"fmt"
	"io"
	"unicode"

	"src.elv.sh/cmd/examples/e3bc/bc"
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

// A highlighter for bc code. Currently this just makes all digits green.
//
// TODO: Highlight more syntax of bc.
type highlighter struct{}

func (highlighter) Get(code string) (ui.Text, []error) {
	t := ui.Text{}
	for _, r := range code {
		var style ui.Styling
		if unicode.IsDigit(r) {
			style = ui.FgGreen
		}
		t = append(t, ui.T(string(r), style)...)
	}
	return t, nil
}

func (highlighter) LateUpdates() <-chan struct{} { return nil }

func main() {
	var app cli.App
	app = cli.NewApp(cli.AppSpec{
		Prompt:      cli.NewConstPrompt(ui.T("bc> ")),
		Highlighter: highlighter{},
		CodeAreaBindings: tk.MapBindings{
			term.K('D', ui.Ctrl): func(tk.Widget) { app.CommitEOF() },
		},
	})

	bc := bc.Start()
	defer bc.Quit()

	for {
		code, err := app.ReadCode()
		if err != nil {
			if err != io.EOF {
				fmt.Println("error:", err)
			}
			break
		}
		bc.Exec(code)
	}
}
