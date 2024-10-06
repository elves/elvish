package main

import (
	"strconv"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

func Counter(c etk.Context) (etk.View, etk.React) {
	valueVar := etk.State(c, "value", 0)
	valueText := ui.T(strconv.Itoa(valueVar.Get()))
	return etk.TextView(1, valueText),
		func(ev term.Event) etk.Reaction {
			switch ev {
			case term.K(ui.Enter):
				valueVar.Swap(func(i int) int { return i + 1 })
				return etk.Consumed
			case term.K('[', ui.Ctrl):
				return etk.Finish
			}
			return etk.Unused
		}
}

func CounterWithButton(c etk.Context) (etk.View, etk.React) {
	valueVar := etk.State(c, "value", 0)
	buttonView, buttonReact := c.Subcomp("button", etk.WithInit(Button,
		"label", "Count",
		"submit", func() { valueVar.Swap(func(i int) int { return i + 1 }) },
	))

	return etk.HBoxFlexView(
			1, 1,
			etk.TextView(0, ui.T(strconv.Itoa(valueVar.Get()))),
			buttonView,
		),
		buttonReact
}

func Button(c etk.Context) (etk.View, etk.React) {
	labelVar := etk.State(c, "label", "button")
	submitVar := etk.State(c, "submit", func() {})
	return etk.TextView(1, ui.T("[ "+labelVar.Get()+" ]")),
		c.WithBinding(func(ev term.Event) etk.Reaction {
			if ev == term.K(' ') || ev == term.K(ui.Enter) {
				submitVar.Get()()
				return etk.Consumed
			}
			return etk.Unused
		})
}
