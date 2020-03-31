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
	vals.TestValue(t, K(-1)).Repr("(edit:key '(bad function key -1)')")
	vals.TestValue(t, K(-2000)).Repr("(edit:key '(bad function key -2000)')")
}

var parseKeyTests = []struct {
	s       string
	wantKey Key
	wantErr string
}{
	{s: "x", wantKey: K('x')},
	{s: "Tab", wantKey: K(Tab)},
	{s: "F1", wantKey: K(F1)},

	// Alt- keys are case-sensitive.
	{s: "a-x", wantKey: Key{'x', Alt}},
	{s: "a-X", wantKey: Key{'X', Alt}},

	// Ctrl- keys are case-insensitive.
	{s: "C-x", wantKey: Key{'X', Ctrl}},
	{s: "C-X", wantKey: Key{'X', Ctrl}},

	// + is the same as -.
	{s: "C+X", wantKey: Key{'X', Ctrl}},

	// Full names and alternative names can also be used.
	{s: "M-x", wantKey: Key{'x', Alt}},
	{s: "Meta-x", wantKey: Key{'x', Alt}},

	// Multiple modifiers can appear in any order.
	{s: "Alt-Ctrl-Delete", wantKey: Key{Delete, Alt | Ctrl}},
	{s: "Ctrl-Alt-Delete", wantKey: Key{Delete, Alt | Ctrl}},

	// Ctrl-I and Ctrl-J are normalized to Tab and Enter.
	{s: "Ctrl-I", wantKey: K(Tab)},
	{s: "Ctrl-J", wantKey: K(Enter)},

	// Errors.
	{s: "F123", wantErr: "bad key: F123"},
	{s: "Super-X", wantErr: "bad modifier: super"},
}

func TestParseKey(t *testing.T) {
	for _, test := range parseKeyTests {
		key, err := ParseKey(test.s)
		if key != test.wantKey {
			t.Errorf("ParseKey(%q) => %v, want %v", test.s, key, test.wantKey)
		}
		if test.wantErr == "" {
			if err != nil {
				t.Errorf("ParseKey(%q) => error %v, want nil", test.s, err)
			}
		} else {
			if err == nil || err.Error() != test.wantErr {
				t.Errorf("ParseKey(%q) => error %v, want error with message %q",
					test.s, err, test.wantErr)
			}
		}
	}
}
