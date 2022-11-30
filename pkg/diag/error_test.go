package diag

import (
	"strings"
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestError(t *testing.T) {
	testutil.Set(t, &culpritLineBegin, "<")
	testutil.Set(t, &culpritLineEnd, ">")

	err := &Error{
		Type:    "some error",
		Message: "bad list",
		Context: *contextInParen("[test]", "echo (x)"),
	}

	wantErrorString := "some error: 5-8 in [test]: bad list"
	if gotErrorString := err.Error(); gotErrorString != wantErrorString {
		t.Errorf("Error() -> %q, want %q", gotErrorString, wantErrorString)
	}

	wantRanging := Ranging{From: 5, To: 8}
	if gotRanging := err.Range(); gotRanging != wantRanging {
		t.Errorf("Range() -> %v, want %v", gotRanging, wantRanging)
	}

	wantShow := lines(
		// Type is capitalized in return value of Show
		"Some error: \033[31;1mbad list\033[m",
		"[test], line 1: echo <(x)>",
	)
	if gotShow := err.Show(""); gotShow != wantShow {
		t.Errorf("Show() -> %q, want %q", gotShow, wantShow)
	}
}

func lines(lines ...string) string {
	return strings.Join(lines, "\n")
}
