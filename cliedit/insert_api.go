package cliedit

import (
	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

func initInsert(ev *eval.Evaler, app *clicore.App) eval.Ns {
	abbr := vals.EmptyMap
	app.CodeArea.Abbreviations = makeMapIterator(&abbr)

	binding := emptyBindingMap
	app.CodeArea.OverlayHandler = newMapBinding(app, ev, &binding)

	quotePaste := false
	quotePasteVar := vars.FromPtr(&quotePaste)
	app.CodeArea.QuotePaste = func() bool { return quotePasteVar.Get().(bool) }

	return eval.Ns{
		"abbr":        vars.FromPtr(&abbr),
		"binding":     vars.FromPtr(&binding),
		"quote-paste": quotePasteVar,
	}
}

func makeMapIterator(m *hashmap.Map) func(func(a, b string)) {
	return func(f func(a, b string)) {
		for it := (*m).Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			ks, kok := k.(string)
			vs, vok := v.(string)
			if !kok || !vok {
				continue
			}
			f(ks, vs)
		}
	}
}
