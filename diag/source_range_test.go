package diag

import (
	"strings"
	"testing"
)

var sourceRangeTests = []struct {
	*Context
	indent            string
	wantPPrint        string
	wantPPrintCompact string
}{
	// Single-line culprit
	{parseContext("echo (bad)", "(", ")", true), "_",
		`
[test], line 1:
_echo <(bad)>`[1:],

		`[test], line 1: echo <(bad)>`,
	},
	// Multi-line culprit
	{parseContext("echo (bad\nbad)", "(", ")", true), "_",
		`
[test], line 1-2:
_echo <(bad>
_<bad)>`[1:],
		`
[test], line 1-2: echo <(bad>
_                  <bad)>`[1:],
	},
	// Empty culprit
	{parseContext("echo x", "x", "x", false), "",
		`
[test], line 1:
echo <^>x`[1:],
		"[test], line 1: echo <^>x",
	},
}

func TestContext(t *testing.T) {
	culpritLineBegin = "<"
	culpritLineEnd = ">"
	for i, test := range sourceRangeTests {
		gotPPrint := test.Context.PPrint(test.indent)
		if gotPPrint != test.wantPPrint {
			t.Errorf("test%d.PPrint(%q) = %q, want %q",
				i, test.indent, gotPPrint, test.wantPPrint)
		}
		gotPPrintCompact := test.Context.PPrintCompact(test.indent)
		if gotPPrintCompact != test.wantPPrintCompact {
			t.Errorf("test%d.PPrintCompact(%q) = %q, want %q",
				i, test.indent, gotPPrintCompact, test.wantPPrintCompact)
		}
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
