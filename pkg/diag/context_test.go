package diag

import (
	"strings"
	"testing"
)

var sourceRangeTests = []struct {
	Name    string
	Context *Context
	Indent  string

	WantShow        string
	WantShowCompact string
}{
	{
		Name:    "single-line culprit",
		Context: parseContext("echo (bad)", "(", ")", true),
		Indent:  "_",

		WantShow: lines(
			"[test], line 1:",
			"_echo <(bad)>",
		),
		WantShowCompact: "[test], line 1: echo <(bad)>",
	},
	{
		Name:    "multi-line culprit",
		Context: parseContext("echo (bad\nbad)", "(", ")", true),
		Indent:  "_",

		WantShow: lines(
			"[test], line 1-2:",
			"_echo <(bad>",
			"_<bad)>",
		),
		WantShowCompact: lines(
			"[test], line 1-2: echo <(bad>",
			"_                  <bad)>",
		),
	},
	{
		Name:    "empty culprit",
		Context: parseContext("echo x", "x", "x", false),
		Indent:  "",

		WantShow: lines(
			"[test], line 1:",
			"echo <^>x",
		),
		WantShowCompact: "[test], line 1: echo <^>x",
	},
}

func TestContext(t *testing.T) {
	culpritLineBegin = "<"
	culpritLineEnd = ">"
	for _, test := range sourceRangeTests {
		t.Run(test.Name, func(t *testing.T) {
			gotShow := test.Context.Show(test.Indent)
			if gotShow != test.WantShow {
				t.Errorf("Show() -> %q, want %q", gotShow, test.WantShow)
			}
			gotShowCompact := test.Context.ShowCompact(test.Indent)
			if gotShowCompact != test.WantShowCompact {
				t.Errorf("ShowCompact() -> %q, want %q",
					gotShowCompact, test.WantShowCompact)
			}
		})
	}
}

// Parse a string into a source range, using the first appearance of certain
// texts as start and end positions.
func parseContext(s, starter, ender string, endAfter bool) *Context {
	end := strings.Index(s, ender)
	if endAfter {
		end += len(ender)
	}
	return NewContext("[test]", s, Ranging{From: strings.Index(s, starter), To: end})
}
