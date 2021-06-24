package edit

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/parse"
)

func setupCompleteGetopt(ev *eval.Evaler) {
	ev.AddBuiltin(eval.NsBuilder{}.AddGoFn("", "complete-getopt", completeGetopt).Ns())

	code := `fn complete [@args]{
	           opt-specs = [ [&short=a &long=all &desc="Show all"]
	                         [&short=n &long=name &desc="Set name" &arg-required=$true
	                          &completer= [_]{ put name1 name2 }] ]
	           arg-handlers = [ [_]{ put first1 first2 }
	                            [_]{ put second1 second2 } ... ]
	           complete-getopt $args $opt-specs $arg-handlers
	         }`
	ev.Eval(parse.Source{Name: "[test init]", Code: code}, eval.EvalCfg{})
}

func TestCompleteGetopt2(t *testing.T) {
	TestWithSetup(t, setupCompleteGetopt,
		// Complete argument
		That("complete ''").Puts("first1", "first2"),
		That("complete '' >&-").Throws(eval.ErrNoValueOutput),

		// Complete option
		That("complete -").Puts(
			complexItem{Stem: "-a", Display: "-a (Show all)"},
			complexItem{Stem: "--all", Display: "--all (Show all)"},
			complexItem{Stem: "-n", Display: "-n (Set name)"},
			complexItem{Stem: "--name", Display: "--name (Set name)"}),
		That("complete - >&-").Throws(eval.ErrNoValueOutput),

		// Complte long option
		That("complete --").Puts(
			complexItem{Stem: "--all", Display: "--all (Show all)"},
			complexItem{Stem: "--name", Display: "--name (Set name)"}),
		That("complete -- >&-").Throws(eval.ErrNoValueOutput),

		// Complete argument of short option
		That("complete -n ''").Puts("name1", "name2"),
		That("complete -n '' >&-").Throws(eval.ErrNoValueOutput),

		// Complete argument of long option
		That("complete --name ''").Puts("name1", "name2"),
		That("complete --name '' >&-").Throws(eval.ErrNoValueOutput),

		// Complete (normal) argument after option that doesn't take an argument
		That("complete -a ''").Puts("first1", "first2"),
		That("complete -a '' >&-").Throws(eval.ErrNoValueOutput),

		// Complete second argument
		That("complete arg1 ''").Puts("second1", "second2"),
		That("complete arg1 '' >&-").Throws(eval.ErrNoValueOutput),

		// Complete variadic argument
		That("complete arg1 arg2 ''").Puts("second1", "second2"),
		That("complete arg1 arg2 '' >&-").Throws(eval.ErrNoValueOutput),
	)
}
