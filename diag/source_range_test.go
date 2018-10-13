package diag

import (
	"strings"
	"testing"
)

var sourceRangeTests = []struct {
	*SourceRange
	indent            string
	wantPPrint        string
	wantPPrintCompact string
}{
	// Single-line culprit
	{parseSourceRange("echo (bad)", "(", ")", true), "_",
		`
[test], line 1:
_echo <(bad)>`[1:],

		`[test], line 1: echo <(bad)>`,
	},
	// Multi-line culprit
	{parseSourceRange("echo (bad\nbad)", "(", ")", true), "_",
		`
[test], line 1-2:
_echo <(bad>
_<bad)>`[1:],
		`
[test], line 1-2: echo <(bad>
_                  <bad)>`[1:],
	},
	// Empty culprit
	{parseSourceRange("echo x", "x", "x", false), "",
		`
[test], line 1:
echo <^>x`[1:],
		"[test], line 1: echo <^>x",
	},
}

func TestSourceRange(t *testing.T) {
	culpritLineBegin = "<"
	culpritLineEnd = ">"
	for i, test := range sourceRangeTests {
		gotPPrint := test.SourceRange.PPrint(test.indent)
		if gotPPrint != test.wantPPrint {
			t.Errorf("test%d.PPrint(%q) = %q, want %q",
				i, test.indent, gotPPrint, test.wantPPrint)
		}
		gotPPrintCompact := test.SourceRange.PPrintCompact(test.indent)
		if gotPPrintCompact != test.wantPPrintCompact {
			t.Errorf("test%d.PPrintCompact(%q) = %q, want %q",
				i, test.indent, gotPPrintCompact, test.wantPPrintCompact)
		}
	}
}

// Parse a string into a source range, using the first appearance of certain
// texts as start and end positions.
func parseSourceRange(s, starter, ender string, endAfter bool) *SourceRange {
	end := strings.Index(s, ender)
	if endAfter {
		end += len(ender)
	}
	return NewSourceRange("[test]", s, strings.Index(s, starter), end)
}
