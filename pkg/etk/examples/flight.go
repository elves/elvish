package main

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

func Flight(c etk.Context) (etk.View, etk.React) {
	typeView, typeReact := c.Subcomp("type",
		etk.WithInit(comps.ListBox,
			"items", comps.StringItems("one-way", "return"),
			"multi-column", true))
	outboundView, outboundReact := c.Subcomp("outbound", comps.TextArea)
	inboundView, inboundReact := c.Subcomp("inbound", comps.TextArea)
	_, bookReact := c.Subcomp("book", Button)
	return Form(c,
		FormComp{"Type:     ", typeView, typeReact, false},
		FormComp{"Outbound: ", outboundView, outboundReact, false},
		FormComp{"Inbound:  ", inboundView, inboundReact,
			etk.BindState(c, "type/selected", 0).Get() == 0},
		FormComp{"[ Book ]", etk.EmptyView{}, bookReact, false})
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

	rows := make([]etk.BoxChild, len(comps))
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
			} else if comp.Disabled {
				label = ui.Concat(label, ui.T(comp.Label, ui.FgBrightBlack))
			} else {
				label = ui.Concat(label, ui.T(comp.Label))
			}
		}
		rows[i].View = etk.Box("label= [comp=]", etk.Text(label), comp.View)
	}
	return etk.BoxView{Children: rows, Focus: focus},
		func(ev term.Event) etk.Reaction {
			reaction := comps[focus].React(ev)
			n := len(comps)
			// TODO: Skip disabled component
			if reaction == etk.Unused {
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
					focusVar.Set((focus + n - 1) % n)
				default:
					return etk.Unused
				}
				return etk.Consumed
			} else if reaction == etk.Finish && focus < n-1 {
				focusVar.Set(focus + 1)
				return etk.Consumed
			}
			return reaction
		}
}
