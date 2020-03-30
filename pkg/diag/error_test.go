package diag

import (
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	err := &Error{
		Type:    "some error",
		Message: "bad list",
		Context: *parseContext("echo [x]", "[", "]", true),
	}

	wantErrorString := "some error: 5-8 in [test]: bad list"
	if gotErrorString := err.Error(); gotErrorString != wantErrorString {
		t.Errorf("Error() -> %q, want %q", gotErrorString, wantErrorString)
	}

	wantRanging := Ranging{From: 5, To: 8}
	if gotRanging := err.Range(); gotRanging != wantRanging {
		t.Errorf("Range() -> %v, want %v", gotRanging, wantRanging)
	}

	culpritLineBegin = "<"
	culpritLineEnd = ">"
	wantShow := lines(
		"some error: \033[31;1mbad list\033[m",
		"[test], line 1: echo <[x]>",
	)
	if gotShow := err.Show(""); gotShow != wantShow {
		t.Errorf("Show() -> %q, want %q", gotShow, wantShow)
	}
}

func lines(lines ...string) string {
	return strings.Join(lines, "\n")
}
