package edit

import (
	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

//elvdoc:var abbr
//
// This is a map of abbreviations to expansions. The expansion occurs as soon
// as the final character of the abbreviation is typed. These are known as
// "instant abbreviations". If more than a single abbreviation would match the
// longest one is used.
//
// Examples:
//
// ```elvish
// edit:abbr['||'] = ' | less'
// edit:abbr['>dn'] = ' 2>/dev/null '
// ```
//
// @cf edit:small-word-abbr

//elvdoc:var small-word-abbr
//
// This is a map of small word abbreviations to expansions. A "small word" is
// a contiguous sequence of whitespace, alpha-numeric, or other characters.
// Small word abbreviations are expanded only when a small word boundary is
// detected at the start and end of the abbreviation. The beginning of the
// command line is considered to be a non-small word character regardless of
// the category of the first character in the abbreviation. The small word
// abbreviation can independently begin and end with a char in any of the
// three aforementioned categories.
//
// Small word abbreviations only expand when the most recently typed character
// is at the end of the command line. This means that if you move the cursor
// to earlier in the line and type what would otherwise match a small word
// abbreviation no expansion will occur.
//
// [Instant abbreviations](#editabbr) have higher priority. Which means small
// word abbreviations are only considered for expansion if no instant
// abbreviation is expanded after typing a character.
//
// Examples:
//
// ```elvish
// edit:small-word-abbr['gcm'] = 'git checkout master'
// edit:small-word-abbr['gcp'] = 'git cherry-pick -x'
// edit:small-word-abbr['ll'] = 'ls -ltr'
// edit:small-word-abbr['>dn'] = ' 2>/dev/null'
// ```
//
// @cf edit:abbr

func initInsertAPI(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, ns eval.Ns) {
	abbr := vals.EmptyMap
	abbrVar := vars.FromPtr(&abbr)
	appSpec.Abbreviations = makeMapIterator(abbrVar)

	SmallWordAbbr := vals.EmptyMap
	SmallWordAbbrVar := vars.FromPtr(&SmallWordAbbr)
	appSpec.SmallWordAbbreviations = makeMapIterator(SmallWordAbbrVar)

	binding := newBindingVar(EmptyBindingMap)
	appSpec.OverlayHandler = newMapBinding(nt, ev, binding)

	quotePaste := newBoolVar(false)
	appSpec.QuotePaste = func() bool { return quotePaste.GetRaw().(bool) }

	toggleQuotePaste := func() {
		quotePaste.Set(!quotePaste.Get().(bool))
	}

	ns.Add("abbr", abbrVar)
	ns.Add("small-word-abbr", SmallWordAbbrVar)
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
