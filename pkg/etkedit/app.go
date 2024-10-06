package edit

import (
	"fmt"
	"strconv"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

type addons struct {
	Addons []addon
	NextID int
}

type addon struct {
	Comp    etk.Comp
	Flex    bool
	Dismiss func()
	ID      int
}

// TODO: Movable focus
func app(c etk.Context) (etk.View, etk.React) {
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
		etk.WithInit(comps.TextArea, "highlighter", hlVar.Get().Get))

	addonsVar := etk.State(c, "addons", addons{})
	addons := addonsVar.Get().Addons

	focusVar := etk.State(c, "focus", 0)
	focus := focusVar.Get()

	v := etk.BoxView{Children: make([]etk.BoxChild, len(addons)+1), Focus: focus}
	reacts := make([]etk.React, len(addons)+1)
	v.Children[0].View, reacts[0] = codeView, codeReact
	for i, addon := range addons {
		v.Children[i+1].View, reacts[i+1] = c.Subcomp(strconv.Itoa(addon.ID), addon.Comp)
		v.Children[i+1].Flex = addons[i].Flex
	}

	binding := c.BindingNopDefault()
	return v,
		func(ev term.Event) etk.Reaction {
			if r := binding(ev); r != etk.Unused {
				return r
			}
			reaction := reacts[focus](ev)
			if focus > 0 && reaction == etk.Finish {
				popAddon(c)
				return etk.Consumed
			}
			if reaction == etk.Unused {
				if k, ok := ev.(term.KeyEvent); ok {
					c.AddMsg(ui.T(fmt.Sprintf("Unbound: %s", ui.Key(k))))
				}
			}
			return reaction
		}
}

func pushAddon(c etk.Context, f etk.Comp, flex bool) {
	pushAddonWithDismiss(c, f, flex, nil)
}

func pushAddonWithDismiss(c etk.Context, f etk.Comp, flex bool, dismiss func()) {
	addonsVar := etk.BindState(c, "addons", addons{})
	addonsVar.Swap(func(a addons) addons {
		// TODO: Is the use of append correct here??
		return addons{append(a.Addons, addon{f, flex, dismiss, a.NextID}), a.NextID + 1}
	})
	c.Set("focus", len(addonsVar.Get().Addons))
}

func popAddon(c etk.Context) {
	// TODO: Remove the state associated with the popped addon
	addonsVar := etk.BindState(c, "addons", addons{})
	var dismiss func()
	addonsVar.Swap(func(a addons) addons {
		if len(a.Addons) > 0 {
			dismiss = a.Addons[len(a.Addons)-1].Dismiss
			a.Addons = a.Addons[:len(a.Addons)-1]
		}
		return a
	})
	// This has to be called outside the Swap call since the dismiss function
	// may also mutate states.
	if dismiss != nil {
		dismiss()
	}
	c.Set("focus", len(addonsVar.Get().Addons))
}
