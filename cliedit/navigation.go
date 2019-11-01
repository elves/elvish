package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/navigation"
	"github.com/elves/elvish/eval"
)

//elvdoc:var navigation:binding
//
// Keybinding for the navigation mode.

//elvdoc:fn navigation:start
//
// Start the navigation mode.

func initNavigation(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("navigation",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFn("<edit:navigation>", "start", func() {
			navigation.Start(app, navigation.Config{Binding: binding})
		}))
}
