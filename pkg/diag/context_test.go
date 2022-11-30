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
		Context: contextInParen("[test]", "echo (bad)"),
		Indent:  "_",

		WantShow: lines(
			"[test], line 1:",
			"_echo <(bad)>",
		),
		WantShowCompact: "[test], line 1: echo <(bad)>",
	},
	{
		Name:    "multi-line culprit",
		Context: contextInParen("[test]", "echo (bad\nbad)\nmore"),
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
		Name: "trailing newline in culprit is removed",
		//                             012345678 9
		Context: NewContext("[test]", "echo bad\n", Ranging{5, 9}),
		Indent:  "_",

		WantShow: lines(
			"[test], line 1:",
			"_echo <bad>",
		),
		WantShowCompact: lines(
			"[test], line 1: echo <bad>",
		),
	},
	{
		Name: "empty culprit",
		//                             012345
		Context: NewContext("[test]", "echo x", Ranging{5, 5}),

		WantShow: lines(
			"[test], line 1:",
			"echo <^>x",
		),
		WantShowCompact: "[test], line 1: echo <^>x",
	},
	{
		Name:            "unknown culprit range",
		Context:         NewContext("[test]", "echo", Ranging{-1, -1}),
		WantShow:        "[test], unknown position",
		WantShowCompact: "[test], unknown position",
	},
	{
		Name:            "invalid culprit range",
		Context:         NewContext("[test]", "echo", Ranging{2, 1}),
		WantShow:        "[test], invalid position 2-1",
		WantShowCompact: "[test], invalid position 2-1",
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

// Returns a Context with the given name and source, and a range for the part
// between ( and ).
func contextInParen(name, src string) *Context {
	return NewContext(name, src,
		Ranging{strings.Index(src, "("), strings.Index(src, ")") + 1})
}
