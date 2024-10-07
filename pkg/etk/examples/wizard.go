package main

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

type Task struct {
	Name        string
	Description string
	Code        string
}

type Tasks []Task

func (ts Tasks) Show(i int) ui.Text { return ui.T(ts[i].Name) }

func (ts Tasks) Len() int { return len(ts) }

var tasks = Tasks{
	{"Set up carapace", "Carapace provides a lot of completions.", "sudo brew install carapace"},
	{"Use readline binding", "Keybindings like:\nCtrl-N to next line\nCtrl-P to previous line\nCtrl-F to next character\nCtrl-B to previous character", "use readline-binding"},
}

func Wizard(c etk.Context) (etk.View, etk.React) {
	listView, listReact := c.Subcomp("list", etk.WithInit(comps.ListBox, "items", tasks))
	selectedVar := etk.BindState(c, "list/selected", 0)
	selected := selectedVar.Get()
	description := etk.TextView(0, ui.T(tasks[selected].Description))
	code := etk.TextView(0, ui.T("\n"+tasks[selected].Code))

	return etk.VBoxView(0, etk.HBoxView(0, listView, description), code),
		func(e term.Event) etk.Reaction {
			if e == term.K(ui.Enter) {
				// TODO: Show notification?
				return etk.Consumed
			}
			return listReact(e)
		}
}
