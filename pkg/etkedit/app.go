package edit

import (
	"strconv"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/etk"
)

type Addons struct {
	Addons []Addon
	NextID int
}

type Addon struct {
	Comp etk.Comp
	ID   int
}

// TODO: Movable focus
func App(c etk.Context) (etk.View, etk.React) {
	hlVar := etk.State(c, "highlighter", (*highlight.Highlighter)(nil))
	if hlVar.Get() == nil {
		hl := getHighlighter(c.Frame().Evaler)
		hlVar.Set(hl)
		go func() {
			for {
				select {
				case <-hl.LateUpdates():
					c.Refresh()
				case <-c.FinishChan():
					return
				}
			}
		}()
	}
	codeView, codeReact := c.Subcomp("code",
		etk.WithInit(etk.TextArea, "highlighter", hlVar.Get().Get))

	addonsVar := etk.State(c, "addons", Addons{})
	addons := addonsVar.Get().Addons

	views := make([]etk.View, len(addons)+1)
	reacts := make([]etk.React, len(addons)+1)
	views[0], reacts[0] = codeView, codeReact
	for i, addon := range addons {
		views[i+1], reacts[i+1] = c.Subcomp(strconv.Itoa(addon.ID), addon.Comp)
	}

	focusVar := etk.State(c, "focus", 0)
	focus := focusVar.Get()

	return etk.VBoxView(focus, views...),
		c.WithBinding(func(ev term.Event) etk.Reaction {
			reaction := reacts[focus](ev)
			if focus > 0 && reaction == etk.Finish {
				PopAddon(c)
				return etk.Consumed
			}
			return reaction
		})
}

func PushAddon(c etk.Context, f etk.Comp) {
	addonsVar := etk.BindState(c, "addons", Addons{})
	addonsVar.Swap(func(a Addons) Addons {
		// TODO: Is the use of append correct here??
		return Addons{append(a.Addons, Addon{f, a.NextID}), a.NextID + 1}
	})
	etk.BindState(c, "focus", 0).Set(len(addonsVar.Get().Addons))
}

func PopAddon(c etk.Context) {
	// TODO: Remove the state associated with the popped addon
	addonsVar := etk.BindState(c, "addons", Addons{})
	addonsVar.Swap(func(a Addons) Addons {
		if len(a.Addons) > 0 {
			a.Addons = a.Addons[:len(a.Addons)-1]
		}
		return a
	})
	etk.BindState(c, "focus", 0).Set(len(addonsVar.Get().Addons))
}
