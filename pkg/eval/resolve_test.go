package eval

import (
	"reflect"
	"sort"
	"testing"

	"github.com/elves/elvish/pkg/eval/vars"
)

var testVar = vars.NewReadOnly("")

var eachVariableInTopTests = []struct {
	builtin   Ns
	global    Ns
	ns        string
	wantNames []string
}{
	{
		builtin:   Ns{"foo": testVar, "bar": testVar},
		global:    Ns{"lorem": testVar, "ipsum": testVar},
		ns:        "builtin:",
		wantNames: []string{"bar", "foo"},
	},
	{
		builtin:   Ns{"foo": testVar, "bar": testVar},
		global:    Ns{"lorem": testVar, "ipsum": testVar},
		ns:        "",
		wantNames: []string{"bar", "foo", "ipsum", "lorem"},
	},
	{
		builtin:   Ns{"mod:": vars.NewReadOnly(Ns{"a": testVar, "b": testVar})},
		ns:        "mod:",
		wantNames: []string{"a", "b"},
	},
	{
		global:    Ns{"mod:": vars.NewReadOnly(Ns{"a": testVar, "b": testVar})},
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
		scopes := evalerScopes{Builtin: test.builtin, Global: test.global}
		fillScopes(&scopes)

		var names []string
		scopes.EachVariableInTop(test.ns,
			func(s string) { names = append(names, s) })
		sort.Strings(names)

		if !reflect.DeepEqual(names, test.wantNames) {
			t.Errorf("got names %v, want %v", names, test.wantNames)
		}
	}
}

var eachNsInTopTests = []struct {
	builtin   Ns
	global    Ns
	wantNames []string
}{
	{
		wantNames: []string{"E:", "builtin:", "e:"},
	},
	{
		builtin:   Ns{"foo:": testVar},
		wantNames: []string{"E:", "builtin:", "e:", "foo:"},
	},
	{
		global:    Ns{"foo:": testVar},
		wantNames: []string{"E:", "builtin:", "e:", "foo:"},
	},
	{
		builtin:   Ns{"foo:": testVar},
		global:    Ns{"bar:": testVar},
		wantNames: []string{"E:", "bar:", "builtin:", "e:", "foo:"},
	},
}

func TestEachNsInTop(t *testing.T) {
	for _, test := range eachNsInTopTests {
		scopes := evalerScopes{Builtin: test.builtin, Global: test.global}
		fillScopes(&scopes)

		var names []string
		scopes.EachNsInTop(func(s string) { names = append(names, s) })
		sort.Strings(names)

		if !reflect.DeepEqual(names, test.wantNames) {
			t.Errorf("got names %v, want %v", names, test.wantNames)
		}
	}
}

func fillScopes(scopes *evalerScopes) {
	if scopes.Builtin == nil {
		scopes.Builtin = Ns{}
	}
	if scopes.Global == nil {
		scopes.Global = Ns{}
	}
}
