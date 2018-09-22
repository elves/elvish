package utils

import (
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/types"
)

var basicHandlerTests = []struct {
	name     string
	keys     []ui.Key
	wantCode string
	wantDot  int
}{
	{"ASCII characters",
		[]ui.Key{{Rune: 'a'}}, "a", 1},
	{"Unicode characters",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}},
		"代码", 6},
	// Backspace
	{"Backspace",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Backspace}},
		"代", 3},
	{"Left 1",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Left}},
		"代码", 3},
	{"Left 2",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Left}, {Rune: ui.Left}},
		"代码", 0},
	{"Right",
		[]ui.Key{{Rune: '代'}, {Rune: '码'}, {Rune: ui.Left}, {Rune: ui.Left},
			{Rune: ui.Right}},
		"代码", 3},
}

func TestBasicHandler(t *testing.T) {
	for _, test := range basicHandlerTests {
		t.Run(test.name, func(t *testing.T) {
			st := types.State{}
			for _, key := range test.keys {
				BasicHandler(key, &st)
			}
			code, dot := st.CodeAndDot()
			if code != test.wantCode {
				t.Errorf("Got code = %q, want %q", code, test.wantCode)
			}
			if dot != test.wantDot {
				t.Errorf("Got dot = %v, want %v", dot, test.wantDot)
			}
		})
	}
}
