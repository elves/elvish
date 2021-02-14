package edit

import (
	"github.com/xiaq/persistent/hashmap"
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

//elvdoc:var abbr
//
// A map from (simple) abbreviations to their expansions.
//
// An abbreviation is replaced by its expansion when it is typed in full
// and consecutively, without being interrupted by the use of other editing
// functionalities, such as cursor movements.
//
// If more than one abbreviations would match, the longest one is used.
//
// Examples:
//
// ```elvish
// edit:abbr['||'] = '| less'
// edit:abbr['>dn'] = '2>/dev/null'
// ```
//
// With the definitions above, typing `||` anywhere expands to `| less`, and
// typing `>dn` anywhere expands to `2>/dev/null`. However, typing a `|`, moving
// the cursor left, and typing another `|` does **not** expand to `| less`,
// since the abbreviation `||` was not typed consecutively.
//
// @cf edit:small-word-abbr

//elvdoc:var small-word-abbr
//
// A map from small-word abbreviations and their expansions.
//
// A small-word abbreviation is replaced by its expansion after it is typed in
// full and consecutively, and followed by another character (the *trigger*
// character). Furthermore, the expansion requires the following conditions to
// be satisfied:
//
// -   The end of the abbreviation must be adjacent to a small-word boundary,
//     i.e. the last character of the abbreviation and the trigger character
//     must be from two different small-word categories.
//
// -   The start of the abbreviation must also be adjacent to a small-word
//     boundary, unless it appears at the beginning of the code buffer.
//
// -   The cursor must be at the end of the buffer.
//
// If more than one abbreviations would match, the longest one is used.
//
// As an example, with the following configuration:
//
// ```elvish
// edit:small-word-abbr['gcm'] = 'git checkout master'
// ```
//
// In the following scenarios, the `gcm` abbreviation is expanded:
//
// -   With an empty buffer, typing `gcm` and a space or semicolon;
//
// -   When the buffer ends with a space, typing `gcm` and a space or semicolon.
//
// The space or semicolon after `gcm` is preserved in both cases.
//
// In the following scenarios, the `gcm` abbreviation is **not** expanded:
//
// -   With an empty buffer, typing `Xgcm` and a space or semicolon (start of
//     abbreviation is not adjacent to a small-word boundary);
//
// -   When the buffer ends with `X`, typing `gcm` and a space or semicolon (end
//     of abbreviation is not adjacent to a small-word boundary);
//
// -   When the buffer is non-empty, move the cursor to the beginning, and typing
//     `gcm` and a space (cursor not at the end of the buffer).
//
// This example shows the case where the abbreviation consists of a single small
// word of alphanumerical characters, but that doesn't have to be the case. For
// example, with the following configuration:
//
// ```elvish
// edit:small-word-abbr['>dn'] = ' 2>/dev/null'
// ```
//
// The abbreviation `>dn` starts with a punctuation character, and ends with an
// alphanumerical character. This means that it is expanded when it borders
// a whitespace or alphanumerical character to the left, and a whitespace or
// punctuation to the right; for example, typing `ls>dn;` will expand it.
//
// Some extra examples of small-word abbreviations:
//
// ```elvish
// edit:small-word-abbr['gcp'] = 'git cherry-pick -x'
// edit:small-word-abbr['ll'] = 'ls -ltr'
// ```
//
// If both a [simple abbreviation](#editabbr) and a small-word abbreviation can
// be expanded, the simple abbreviation has priority.
//
// @cf edit:abbr

func initInsertAPI(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, nb eval.NsBuilder) {
	abbr := vals.EmptyMap
	abbrVar := vars.FromPtr(&abbr)
	appSpec.Abbreviations = makeMapIterator(abbrVar)

	SmallWordAbbr := vals.EmptyMap
	SmallWordAbbrVar := vars.FromPtr(&SmallWordAbbr)
	appSpec.SmallWordAbbreviations = makeMapIterator(SmallWordAbbrVar)

	bindingVar := newBindingVar(emptyBindingsMap)
	appSpec.CodeAreaBindings = newMapBindings(nt, ev, bindingVar)

	quotePaste := newBoolVar(false)
	appSpec.QuotePaste = func() bool { return quotePaste.GetRaw().(bool) }

	toggleQuotePaste := func() {
		quotePaste.Set(!quotePaste.Get().(bool))
	}

	nb.Add("abbr", abbrVar)
	nb.Add("small-word-abbr", SmallWordAbbrVar)
	nb.AddGoFn("<edit>", "toggle-quote-paste", toggleQuotePaste)
	nb.AddNs("insert", eval.NsBuilder{
		"binding":     bindingVar,
		"quote-paste": quotePaste,
	}.Ns())
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
