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
	builtin   *eval.Ns
	global    *eval.Ns
	ns        string
	wantNames []string
}{
	{
		builtin:   eval.NsBuilder{"foo": testVar, "bar": testVar}.Ns(),
		global:    eval.NsBuilder{"lorem": testVar, "ipsum": testVar}.Ns(),
		ns:        "",
		wantNames: []string{"bar", "foo", "ipsum", "lorem"},
	},
	{
		builtin: eval.NsBuilder{
			"mod:": vars.NewReadOnly(eval.NsBuilder{"a": testVar, "b": testVar}.Ns()),
		}.Ns(),
		ns:        "mod:",
		wantNames: []string{"a", "b"},
	},
	{
		global: eval.NsBuilder{
			"mod:": vars.NewReadOnly(eval.NsBuilder{"a": testVar, "b": testVar}.Ns()),
		}.Ns(),
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
	builtin   *eval.Ns
	global    *eval.Ns
	wantNames []string
}{
	{
		wantNames: []string{"E:", "e:"},
	},
	{
		builtin:   eval.NsBuilder{"foo:": testVar}.Ns(),
		wantNames: []string{"E:", "e:", "foo:"},
	},
	{
		global:    eval.NsBuilder{"foo:": testVar}.Ns(),
		wantNames: []string{"E:", "e:", "foo:"},
	},
	{
		builtin:   eval.NsBuilder{"foo:": testVar}.Ns(),
		global:    eval.NsBuilder{"bar:": testVar}.Ns(),
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

func getNs(ns *eval.Ns) *eval.Ns {
	if ns == nil {
		return new(eval.Ns)
	}
	return ns
}
