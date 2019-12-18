package edit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

func initInsertAPI(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, ns eval.Ns) {
	abbr := vals.EmptyMap
	abbrVar := vars.FromPtr(&abbr)
	appSpec.Abbreviations = makeMapIterator(abbrVar)

	binding := newBindingVar(EmptyBindingMap)
	appSpec.OverlayHandler = newMapBinding(nt, ev, binding)

	quotePaste := newBoolVar(false)
	appSpec.QuotePaste = func() bool { return quotePaste.GetRaw().(bool) }

	toggleQuotePaste := func() {
		quotePaste.Set(!quotePaste.Get().(bool))
	}

	ns.Add("abbr", abbrVar)
	ns.AddGoFn("<edit>", "toggle-quote-paste", toggleQuotePaste)
	ns.AddNs("insert", eval.Ns{
		"binding":     binding,
		"quote-paste": quotePaste,
	})
}

func makeMapIterator(mv vars.PtrVar) func(func(a, b string)) {
	return func(f func(a, b string)) {
		for it := mv.GetRaw().(hashmap.Map).Iterator(); it.HasElem(); it.Next() {
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
