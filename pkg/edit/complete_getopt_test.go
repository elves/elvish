package edit

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
)

func TestCompleteGetopt(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`fn complete [@args]{
		   opt-specs = [ [&short=a &long=all &desc="Show all"]
		                 [&short=n &long=name &desc="Set name" &arg-required=$true
		                  &completer= [_]{ put name1 name2 }] ]
		   arg-handlers = [ [_]{ put first1 first2 }
		                    [_]{ put second1 second2 } ... ]
		   edit:complete-getopt $args $opt-specs $arg-handlers
		 }`,
		`@arg1 = (complete '')`,
		`@opts = (complete -)`,
		`@long-opts = (complete --)`,
		`@short-opt-arg = (complete -n '')`,
		`@long-opt-arg = (complete --name '')`,
		`@arg1-after-opt = (complete -a '')`,
		`@arg2 = (complete arg1 '')`,
		`@vararg = (complete arg1 arg2 '')`,
	)
	testGlobals(t, f.Evaler, map[string]interface{}{
		"arg1": vals.MakeList("first1", "first2"),
		"opts": vals.MakeList(
			complexItem{Stem: "-a", Display: "-a (Show all)"},
			complexItem{Stem: "--all", Display: "--all (Show all)"},
			complexItem{Stem: "-n", Display: "-n (Set name)"},
			complexItem{Stem: "--name", Display: "--name (Set name)"}),
		"long-opts": vals.MakeList(
			complexItem{Stem: "--all", Display: "--all (Show all)"},
			complexItem{Stem: "--name", Display: "--name (Set name)"}),
		"short-opt-arg":  vals.MakeList("name1", "name2"),
		"long-opt-arg":   vals.MakeList("name1", "name2"),
		"arg1-after-opt": vals.MakeList("first1", "first2"),
		"arg2":           vals.MakeList("second1", "second2"),
		"vararg":         vals.MakeList("second1", "second2"),
	})
}
