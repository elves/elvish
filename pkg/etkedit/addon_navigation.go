package edit

import "src.elv.sh/pkg/etk"

func startNavigation(ed *Editor, c etk.Context) {
	pushAddon(c, etk.WithInit(
		navigation,
		"binding", etkBindingFromBindingMap(ed, &ed.navigationBinding)), 1)
}

func navigation(c etk.Context) (etk.View, etk.React) {
	return nil, nil
}
