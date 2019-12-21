package diag

import (
	"strings"
	"testing"
)

var sourceRangeTests = []struct {
	Name    string
	Context *Context
	Indent  string

	WantPPrint        string
	WantPPrintCompact string
}{
	{
		Name:    "single-line culprit",
		Context: parseContext("echo (bad)", "(", ")", true),
		Indent:  "_",

		WantPPrint: lines(
			"[test], line 1:",
			"_echo <(bad)>",
		),
		WantPPrintCompact: "[test], line 1: echo <(bad)>",
	},
	{
		Name:    "multi-line culprit",
		Context: parseContext("echo (bad\nbad)", "(", ")", true),
		Indent:  "_",

		WantPPrint: lines(
			"[test], line 1-2:",
			"_echo <(bad>",
			"_<bad)>",
		),
		WantPPrintCompact: lines(
			"[test], line 1-2: echo <(bad>",
			"_                  <bad)>",
		),
	},
	{
		Name:    "empty culprit",
		Context: parseContext("echo x", "x", "x", false),
		Indent:  "",

		WantPPrint: lines(
			"[test], line 1:",
			"echo <^>x",
		),
		WantPPrintCompact: "[test], line 1: echo <^>x",
	},
}

func TestContext(t *testing.T) {
	culpritLineBegin = "<"
	culpritLineEnd = ">"
	for _, test := range sourceRangeTests {
		t.Run(test.Name, func(t *testing.T) {
			gotPPrint := test.Context.PPrint(test.Indent)
			if gotPPrint != test.WantPPrint {
				t.Errorf("PPrint() -> %q, want %q", gotPPrint, test.WantPPrint)
			}
			gotPPrintCompact := test.Context.PPrintCompact(test.Indent)
			if gotPPrintCompact != test.WantPPrintCompact {
				t.Errorf("PPrintCompact() -> %q, want %q",
					gotPPrintCompact, test.WantPPrintCompact)
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
	return NewContext("[test]", s, strings.Index(s, starter), end)
}
