package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

//elvdoc:var abbr
//
// A map from simple abbreviations to their expansions.
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
// set edit:abbr['||'] = '| less'
// set edit:abbr['>dn'] = '2>/dev/null'
// ```
//
// With the definitions above, typing `||` anywhere expands to `| less`, and
// typing `>dn` anywhere expands to `2>/dev/null`. However, typing a `|`, moving
// the cursor left, and typing another `|` does **not** expand to `| less`,
// since the abbreviation `||` was not typed consecutively.
//
// @cf edit:command-abbr edit:small-word-abbr

//elvdoc:var command-abbr
//
// A map from command abbreviations to their expansions.
//
// A command abbreviation is replaced by its expansion when seen in the command
// position followed by a [whitespace](language.html#whitespace). This is
// similar to the Fish shell's
// [abbreviations](https://fishshell.com/docs/current/cmds/abbr.html), but does
// not trigger when executing a command with Enter -- you must type a space
// first.
//
// Examples:
//
// ```elvish
// set edit:command-abbr['l'] = 'less'
// set edit:command-abbr['gc'] = 'git commit'
// ```
//
// @cf edit:abbr edit:small-word-abbr

//elvdoc:var small-word-abbr
//
// A map from small-word abbreviations to their expansions.
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
// If more than one abbreviations would match, the longest one is used. See the description of
// [small words](#word-types) for more information.
//
// As an example, with the following configuration:
//
// ```elvish
// set edit:small-word-abbr['gcm'] = 'git checkout master'
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
// set edit:small-word-abbr['>dn'] = ' 2>/dev/null'
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
// set edit:small-word-abbr['gcp'] = 'git cherry-pick -x'
// set edit:small-word-abbr['ll'] = 'ls -ltr'
// ```
//
// If both a [simple abbreviation](#edit:abbr) and a small-word abbreviation can
// be expanded, the simple abbreviation has priority.
//
// @cf edit:abbr edit:command-abbr

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
