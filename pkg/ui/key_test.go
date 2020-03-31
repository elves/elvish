package ui

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/xiaq/persistent/hash"
)

var kTests = []struct {
	k1 Key
	k2 Key
}{
	{K('a'), Key{'a', 0}},
	{K('a', Alt), Key{'a', Alt}},
	{K('a', Alt, Ctrl), Key{'a', Alt | Ctrl}},
}

func TestK(t *testing.T) {
	for _, test := range kTests {
		if test.k1 != test.k2 {
			t.Errorf("%v != %v", test.k1, test.k2)
		}
	}
}

func TestKeyAsElvishValue(t *testing.T) {
	vals.TestValue(t, K('a')).
		Kind("edit:key").
		Hash(hash.DJB('a', 0)).
		Repr("(edit:key a)").
		Equal(K('a')).
		NotEqual(K('A'), K('a', Alt))

	vals.TestValue(t, K('a', Alt)).Repr("(edit:key Alt-a)")
	vals.TestValue(t, K('a', Ctrl, Alt, Shift)).Repr("(edit:key Ctrl-Alt-Shift-a)")

	vals.TestValue(t, K('\t')).Repr("(edit:key Tab)")
	vals.TestValue(t, K(F1)).Repr("(edit:key F1)")
}

var parseKeyTests = []struct {
	s       string
	wantKey Key
}{
	// Alt- keys are case-sensitive.
	{"a-x", Key{'x', Alt}},
	{"a-X", Key{'X', Alt}},

	// Ctrl- keys are case-insensitive.
	{"C-x", Key{'X', Ctrl}},
	{"C-X", Key{'X', Ctrl}},

	// + is the same as -.
	{"C+X", Key{'X', Ctrl}},

	// Full names and alternative names can also be used.
	{"M-x", Key{'x', Alt}},
	{"Meta-x", Key{'x', Alt}},

	// Multiple modifiers can appear in any order.
	{"Alt-Ctrl-Delete", Key{Delete, Alt | Ctrl}},
	{"Ctrl-Alt-Delete", Key{Delete, Alt | Ctrl}},

	// Ctrl-I and Ctrl-J are normalized to Tab and Enter.
	{"Ctrl-I", K(Tab)},
	{"Ctrl-J", K(Enter)},
}

func TestParseKey(t *testing.T) {
	for _, test := range parseKeyTests {
		key, err := parseKey(test.s)
		if key != test.wantKey {
			t.Errorf("ParseKey(%q) => %v, want %v", test.s, key, test.wantKey)
		}
		if err != nil {
			t.Errorf("ParseKey(%q) => error %v, want nil", test.s, err)
		}
	}
}
