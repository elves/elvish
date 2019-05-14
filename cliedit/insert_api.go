package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

func initInsert(ev *eval.Evaler, cfg *cli.InsertModeConfig) eval.Ns {
	// Underlying abbreviation map and binding map.
	abbr := vals.EmptyMap
	binding := emptyBindingMap

	cfg.Binding = newMapBinding(ev, &binding)
	cfg.Abbrs = newMapStringPairs(&abbr)

	ns := eval.Ns{
		"binding":     vars.FromPtr(&binding),
		"abbr":        vars.FromPtr(&abbr),
		"quote-paste": vars.FromPtr(&cfg.QuotePaste),
	}.AddGoFns("<edit:insert>:", map[string]interface{}{
		"start":           cli.StartInsert,
		"default-handler": cli.DefaultInsert,
	})

	return ns
}

func newMapStringPairs(m *hashmap.Map) cli.StringPairs {
	return mapStringPairs{m}
}

type mapStringPairs struct{ m *hashmap.Map }

func (s mapStringPairs) IterateStringPairs(f func(a, b string)) {
	for it := (*s.m).Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		ks, kok := k.(string)
		vs, vok := v.(string)
		if !kok || !vok {
			continue
		}
		f(ks, vs)
	}
}
