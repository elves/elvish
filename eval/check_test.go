package eval

import (
	"reflect"
	"testing"
)

func prepareChecker(text string) *Checker {
	ch := NewChecker()
	ch.startCheck("<test>", text, map[string]Type{})
	return ch
}

func TestScope(t *testing.T) {
	ch := prepareChecker("")
	assertScopes := func(when string, wanted []map[string]Type) {
		if !reflect.DeepEqual(ch.scopes, wanted) {
			t.Errorf("%s, ch.scopes = %v, want %v", when, ch.scopes, wanted)
		}
	}
	assertHasVarOnThisScope := func(name string, wanted bool) {
		if got := ch.hasVarOnThisScope(name); got != wanted {
			t.Errorf("ch.hasVarOnThisScope(%q) = %v, wanted %v", name, got, wanted)
		}
	}
	empty := map[string]Type{}
	assertScopes("at init", []map[string]Type{empty})

	ch.pushScope()
	ch.pushVar("foo", StringType{})
	assertScopes("after pushing", []map[string]Type{
		empty,
		map[string]Type{"foo": StringType{}},
	})

	ch.pushScope()
	ch.pushVar("bar", TableType{})
	assertScopes("after two pushings", []map[string]Type{
		empty,
		map[string]Type{"foo": StringType{}},
		map[string]Type{"bar": TableType{}},
	})

	assertHasVarOnThisScope("mua", false)
	assertHasVarOnThisScope("foo", false)
	assertHasVarOnThisScope("bar", true)

	ch.popScope()
	assertScopes("after pushing", []map[string]Type{
		empty,
		map[string]Type{"foo": StringType{}},
	})
}
