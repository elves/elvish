package edit

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

func setupCompleteGetopt(ev *eval.Evaler) {
	ev.ExtendBuiltin(eval.BuildNs().AddGoFn("complete-getopt", completeGetopt))

	code := `fn complete {|@args|
	           var opt-specs = [ [&short=a &long=all &desc="Show all"]
	                             [&short=n &long=name &desc="Set name"
	                              &arg-required=$true &arg-desc='new-name'
	                              &completer= {|_| put name1 name2 }] ]
	           var arg-handlers = [ {|_| put first1 first2 }
	                                {|_| put second1 second2 } ... ]
	           complete-getopt $args $opt-specs $arg-handlers
	         }`
	ev.Eval(parse.Source{Name: "[test init]", Code: code}, eval.EvalCfg{})
}

func TestCompleteGetopt(t *testing.T) {
	TestWithEvalerSetup(t, setupCompleteGetopt,
		// Complete argument
		That("complete ''").Puts("first1", "first2"),
		That("complete '' >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete option
		That("complete -").Puts(
			complexItem{Stem: "-a", Display: ui.T("-a (Show all)")},
			complexItem{Stem: "--all", Display: ui.T("--all (Show all)")},
			complexItem{Stem: "-n", Display: ui.T("-n new-name (Set name)")},
			complexItem{Stem: "--name", Display: ui.T("--name new-name (Set name)")}),
		That("complete - >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete long option
		That("complete --").Puts(
			complexItem{Stem: "--all", Display: ui.T("--all (Show all)")},
			complexItem{Stem: "--name", Display: ui.T("--name new-name (Set name)")}),
		That("complete --a").Puts(
			complexItem{Stem: "--all", Display: ui.T("--all (Show all)")}),
		That("complete -- >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete argument of short option
		That("complete -n ''").Puts("name1", "name2"),
		That("complete -n '' >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete argument of long option
		That("complete --name ''").Puts("name1", "name2"),
		That("complete --name '' >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete (normal) argument after option that doesn't take an argument
		That("complete -a ''").Puts("first1", "first2"),
		That("complete -a '' >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete second argument
		That("complete arg1 ''").Puts("second1", "second2"),
		That("complete arg1 '' >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),

		// Complete variadic argument
		That("complete arg1 arg2 ''").Puts("second1", "second2"),
		That("complete arg1 arg2 '' >&-").Throws(eval.ErrPortDoesNotSupportValueOutput),
	)
}

func TestCompleteGetopt_TypeCheck(t *testing.T) {
	TestWithEvalerSetup(t, setupCompleteGetopt,
		That("complete-getopt [foo []] [] []").
			Throws(ErrorWithMessage("arg should be string, got list")),

		That("complete-getopt [] [foo] []").
			Throws(ErrorWithMessage("opt should be map, got string")),
		That("complete-getopt [] [[&short=[]]] []").
			Throws(ErrorWithMessage("short should be string, got list")),
		That("complete-getopt [] [[&short=foo]] []").
			Throws(ErrorWithMessage("short should be exactly one rune, got foo")),
		That("complete-getopt [] [[&long=[]]] []").
			Throws(ErrorWithMessage("long should be string, got list")),
		That("complete-getopt [] [[&]] []").
			Throws(ErrorWithMessage("opt should have at least one of short and long forms")),
		That("complete-getopt [] [[&short=a &arg-required=foo]] []").
			Throws(ErrorWithMessage("arg-required should be bool, got string")),
		That("complete-getopt [] [[&short=a &arg-optional=foo]] []").
			Throws(ErrorWithMessage("arg-optional should be bool, got string")),
		That("complete-getopt [] [[&short=a &arg-required=$true &arg-optional=$true]] []").
			Throws(ErrorWithMessage("opt cannot have both arg-required and arg-optional")),
		That("complete-getopt [] [[&short=a &desc=[]]] []").
			Throws(ErrorWithMessage("desc should be string, got list")),
		That("complete-getopt [] [[&short=a &arg-desc=[]]] []").
			Throws(ErrorWithMessage("arg-desc should be string, got list")),
		That("complete-getopt [] [[&short=a &completer=[]]] []").
			Throws(ErrorWithMessage("completer should be fn, got list")),

		That("complete-getopt [] [] [foo]").
			Throws(ErrorWithMessage("string except for ... not allowed as argument handler, got foo")),
		That("complete-getopt [] [] [[]]").
			Throws(ErrorWithMessage("argument handler should be fn, got list")),
	)
}
