package main

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

func Flight(c etk.Context) (etk.View, etk.React) {
	typeView, typeReact := c.Subcomp("type",
		etk.WithInit(etk.ListBox,
			"items", etk.StringItems("one-way", "return"),
			"horizontal", true))
	outboundView, outboundReact := c.Subcomp("outbound", etk.TextArea)
	// TODO: Disable inbound for one-way
	inboundView, inboundReact := c.Subcomp("inbound", etk.TextArea)
	bookView, bookReact := c.Subcomp("book", etk.WithInit(Button, "label", "Book"))
	return Form(c,
		FormComp{"Type:     ", typeView, typeReact, false},
		FormComp{"Outbound: ", outboundView, outboundReact, false},
		FormComp{"Inbound:  ", inboundView, inboundReact, false},
		FormComp{"", bookView, bookReact, false})
}

type FormComp struct {
	Label    string
	View     etk.View
	React    etk.React
	Disabled bool
}

func Form(c etk.Context, comps ...FormComp) (etk.View, etk.React) {
	focusVar := etk.State(c, "focus", 0)
	focus := focusVar.Get()

	rows := make([]etk.View, len(comps))
	for i, comp := range comps {
		var label ui.Text
		if i == focus {
			label = ui.T("> ", ui.Bold, ui.FgGreen)
		} else {
			label = ui.T("  ")
		}
		if comp.Label != "" {
			if i == focus {
				label = ui.Concat(label, ui.T(comp.Label, ui.Bold))
			} else {
				label = ui.Concat(label, ui.T(comp.Label))
			}
		}
		rows[i] = etk.HBoxFlexView(1, 0, etk.TextView(1, label), comp.View)
	}
	return etk.VBoxView(focus, rows...), func(ev term.Event) etk.Reaction {
		reaction := comps[focus].React(ev)
		if reaction == etk.Unused {
			n := len(comps)
			switch ev {
			case term.K(ui.Down), term.K(ui.Enter):
				if focus < n-1 {
					focusVar.Set(focus + 1)
				}
			case term.K(ui.Up):
				if focus > 0 {
					focusVar.Set(focus - 1)
				}
			case term.K(ui.Tab):
				focusVar.Set((focus + 1) % n)
			case term.K(ui.Tab, ui.Shift):
				focusVar.Set((focus - 1) % n)
			}
		}
		return reaction
	}
}
