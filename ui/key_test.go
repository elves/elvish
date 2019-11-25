package ui

import "testing"

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
