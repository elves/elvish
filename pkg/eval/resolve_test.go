package eval

import (
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/pkg/eval/vars"
)

var testVar = vars.NewReadOnly("")

var eachVariableInTopTests = []struct {
	builtin   *Ns
	global    *Ns
	ns        string
	wantNames []string
}{
	{
		builtin:   NsBuilder{"foo": testVar, "bar": testVar}.Ns(),
		global:    NsBuilder{"lorem": testVar, "ipsum": testVar}.Ns(),
		ns:        "builtin:",
		wantNames: []string{"bar", "foo"},
	},
	{
		builtin:   NsBuilder{"foo": testVar, "bar": testVar}.Ns(),
		global:    NsBuilder{"lorem": testVar, "ipsum": testVar}.Ns(),
		ns:        "",
		wantNames: []string{"bar", "foo", "ipsum", "lorem"},
	},
	{
		builtin: NsBuilder{
			"mod:": vars.NewReadOnly(NsBuilder{"a": testVar, "b": testVar}.Ns()),
		}.Ns(),
		ns:        "mod:",
		wantNames: []string{"a", "b"},
	},
	{
		global: NsBuilder{
			"mod:": vars.NewReadOnly(NsBuilder{"a": testVar, "b": testVar}.Ns()),
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
		EachVariableInTop(builtin, global, test.ns,
			func(s string) { names = append(names, s) })
		sort.Strings(names)

		if !reflect.DeepEqual(names, test.wantNames) {
			t.Errorf("got names %v, want %v", names, test.wantNames)
		}
	}
}

var eachNsInTopTests = []struct {
	builtin   *Ns
	global    *Ns
	wantNames []string
}{
	{
		wantNames: []string{"E:", "builtin:", "e:"},
	},
	{
		builtin:   NsBuilder{"foo:": testVar}.Ns(),
		wantNames: []string{"E:", "builtin:", "e:", "foo:"},
	},
	{
		global:    NsBuilder{"foo:": testVar}.Ns(),
		wantNames: []string{"E:", "builtin:", "e:", "foo:"},
	},
	{
		builtin:   NsBuilder{"foo:": testVar}.Ns(),
		global:    NsBuilder{"bar:": testVar}.Ns(),
		wantNames: []string{"E:", "bar:", "builtin:", "e:", "foo:"},
	},
}

func TestEachNsInTop(t *testing.T) {
	for _, test := range eachNsInTopTests {
		builtin := getNs(test.builtin)
		global := getNs(test.global)

		var names []string
		EachNsInTop(builtin, global, func(s string) { names = append(names, s) })
		sort.Strings(names)

		if !reflect.DeepEqual(names, test.wantNames) {
			t.Errorf("got names %v, want %v", names, test.wantNames)
		}
	}
}

func getNs(ns *Ns) *Ns {
	if ns == nil {
		return new(Ns)
	}
	return ns
}
