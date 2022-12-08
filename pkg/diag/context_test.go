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

		WantShow: dedent(`
			[test]:1:6:
			_  echo <(bad)>`),
		WantShowCompact: "[test]:1:6: echo <(bad)>",
	},
	{
		Name:    "multi-line culprit",
		Context: contextInParen("[test]", "echo (bad\nbad)\nmore"),
		Indent:  "_",

		WantShow: dedent(`
			[test]:1:6:
			_  echo <(bad>
			_  <bad)>`),
		WantShowCompact: dedent(`
			[test]:1:6: echo <(bad>
			_            <bad)>`),
	},
	{
		Name:    "trailing newline in culprit is removed",
		Context: NewContext("[test]", "echo bad\n", Ranging{5, 9}),
		Indent:  "_",

		WantShow: dedent(`
			[test]:1:6:
			_  echo <bad>`),
		WantShowCompact: "[test]:1:6: echo <bad>",
	},
	{
		Name:    "empty culprit",
		Context: NewContext("[test]", "echo x", Ranging{5, 5}),

		WantShow: dedent(`
			[test]:1:6:
			  echo <^>x`),
		WantShowCompact: "[test]:1:6: echo <^>x",
	},
}

func TestContext(t *testing.T) {
	setCulpritMarkers(t, "<", ">")
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
