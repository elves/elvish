package util

import (
	"strings"
	"testing"
)

var sourceRangeTests = []struct {
	*SourceRange
	indent            string
	wantPprint        string
	wantPprintCompact string
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
	CulpritLineBegin = "<"
	CulpritLineEnd = ">"
	for i, test := range sourceRangeTests {
		gotPprint := test.SourceRange.Pprint(test.indent)
		if gotPprint != test.wantPprint {
			t.Errorf("test%d.Pprint(%q) = %q, want %q",
				i, test.indent, gotPprint, test.wantPprint)
		}
		gotPprintCompact := test.SourceRange.PprintCompact(test.indent)
		if gotPprintCompact != test.wantPprintCompact {
			t.Errorf("test%d.PprintCompact(%q) = %q, want %q",
				i, test.indent, gotPprintCompact, test.wantPprintCompact)
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
	return NewSourceRange("[test]", s, strings.Index(s, starter), end, nil)
}
