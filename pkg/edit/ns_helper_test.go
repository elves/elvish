package edit

import (
	"reflect"
	"sort"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
)

var testVar = vars.NewReadOnly("")

var eachVariableInTopTests = []struct {
	builtin   eval.Nser
	global    eval.Nser
	ns        string
	wantNames []string
}{
	{
		builtin:   eval.BuildNs().AddVar("foo", testVar).AddVar("bar", testVar),
		global:    eval.BuildNs().AddVar("lorem", testVar).AddVar("ipsum", testVar),
		ns:        "",
		wantNames: []string{"bar", "foo", "ipsum", "lorem"},
	},
	{
		builtin: eval.BuildNs().AddNs("mod",
			eval.BuildNs().AddVar("a", testVar).AddVar("b", testVar)),
		ns:        "mod:",
		wantNames: []string{"a", "b"},
	},
	{
		global: eval.BuildNs().AddNs("mod",
			eval.BuildNs().AddVar("a", testVar).AddVar("b", testVar)),
		ns:        "mod:",
		wantNames: []string{"a", "b"},
	},
	{
		ns:        "mod:",
		wantNames: nil,
	},
}

func TestEachVariableInTop(t *testing.T) {
	for _, test := range eachVariableInTopTests {
		builtin := getNs(test.builtin)
		global := getNs(test.global)

		var names []string
		eachVariableInTop(builtin, global, test.ns,
			func(s string) { names = append(names, s) })
		sort.Strings(names)

		if !reflect.DeepEqual(names, test.wantNames) {
			t.Errorf("got names %v, want %v", names, test.wantNames)
		}
	}
}

var eachNsInTopTests = []struct {
	builtin   eval.Nser
	global    eval.Nser
	wantNames []string
}{
	{
		wantNames: []string{"E:", "e:"},
	},
	{
		builtin:   eval.BuildNs().AddNs("foo", eval.BuildNs()),
		wantNames: []string{"E:", "e:", "foo:"},
	},
	{
		global:    eval.BuildNs().AddNs("foo", eval.BuildNs()),
		wantNames: []string{"E:", "e:", "foo:"},
	},
	{
		builtin:   eval.BuildNs().AddNs("foo", eval.BuildNs()),
		global:    eval.BuildNs().AddNs("bar", eval.BuildNs()),
		wantNames: []string{"E:", "bar:", "e:", "foo:"},
	},
}

func TestEachNsInTop(t *testing.T) {
	for _, test := range eachNsInTopTests {
		builtin := getNs(test.builtin)
		global := getNs(test.global)

		var names []string
		eachNsInTop(builtin, global, func(s string) { names = append(names, s) })
		sort.Strings(names)

		if !reflect.DeepEqual(names, test.wantNames) {
			t.Errorf("got names %v, want %v", names, test.wantNames)
		}
	}
}

func getNs(ns eval.Nser) *eval.Ns {
	if ns == nil {
		return new(eval.Ns)
	}
	return ns.Ns()
}
