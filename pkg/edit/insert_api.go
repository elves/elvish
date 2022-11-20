package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

func initInsertAPI(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, nb eval.NsBuilder) {
	simpleAbbr := vals.EmptyMap
	simpleAbbrVar := vars.FromPtr(&simpleAbbr)
	appSpec.SimpleAbbreviations = makeMapIterator(simpleAbbrVar)

	commandAbbr := vals.EmptyMap
	commandAbbrVar := vars.FromPtr(&commandAbbr)
	appSpec.CommandAbbreviations = makeMapIterator(commandAbbrVar)

	smallWordAbbr := vals.EmptyMap
	smallWordAbbrVar := vars.FromPtr(&smallWordAbbr)
	appSpec.SmallWordAbbreviations = makeMapIterator(smallWordAbbrVar)

	bindingVar := newBindingVar(emptyBindingsMap)
	appSpec.CodeAreaBindings = newMapBindings(nt, ev, bindingVar)

	quotePaste := newBoolVar(false)
	appSpec.QuotePaste = func() bool { return quotePaste.GetRaw().(bool) }

	toggleQuotePaste := func() {
		quotePaste.Set(!quotePaste.Get().(bool))
	}

	nb.AddVar("abbr", simpleAbbrVar)
	nb.AddVar("command-abbr", commandAbbrVar)
	nb.AddVar("small-word-abbr", smallWordAbbrVar)
	nb.AddGoFn("toggle-quote-paste", toggleQuotePaste)
	nb.AddNs("insert", eval.BuildNs().
		AddVar("binding", bindingVar).
		AddVar("quote-paste", quotePaste))
}

func makeMapIterator(mv vars.PtrVar) func(func(a, b string)) {
	return func(f func(a, b string)) {
		for it := mv.GetRaw().(vals.Map).Iterator(); it.HasElem(); it.Next() {
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
