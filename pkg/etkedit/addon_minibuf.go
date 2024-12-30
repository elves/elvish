package edit

import (
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

// TODO: Binding
func startMinibuf(c etk.Context) {
	pushAddon(c, withFinish(
		etk.WithInit(comps.TextArea, "prompt", addonPromptText(" MINIBUF ")),
		func(c etk.Context) {
			code := etk.BindState(c, "buffer", comps.TextBuffer{}).Get().Content
			src := parse.Source{Name: "[minibuf]", Code: code}
			notifyPort, cleanup := makeNotifyPort(c)
			defer cleanup()
			ports := []*eval.Port{eval.DummyInputPort, notifyPort, notifyPort}
			err := c.Frame().Evaler.Eval(src, eval.EvalCfg{Ports: ports})
			if err != nil {
				c.AddMsg(modes.ErrorText(err))
			}
		},
	), 1)
}

// TODO: Is this the correct abstraction??
//
// This feels a bit like solving the same problem as (etk.Context).WithBinding,
// just from "outside" rather than "inside"?
//
// - WithBinding makes it possible to override from multiple levels higher, but
// can't compose multiple overrides
//
// - This allows composing multiple overrides, but only one level
func withFinish(f etk.Comp, finishFn func(etk.Context)) etk.Comp {
	return withAfterReact(f, func(c etk.Context, r etk.Reaction) etk.Reaction {
		if r == etk.Finish {
			finishFn(c)
		}
		return r
	})
}

func withAfterReact(f etk.Comp, afterFn func(etk.Context, etk.Reaction) etk.Reaction) etk.Comp {
	return func(c etk.Context) (etk.View, etk.React) {
		v, r := f(c)
		return v, func(e term.Event) etk.Reaction {
			reaction := r(e)
			return afterFn(c, reaction)
		}
	}
}
