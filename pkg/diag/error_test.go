package diag

import (
	"testing"
)

func TestError(t *testing.T) {
	setCulpritMarkers(t, "<", ">")
	setMessageMarkers(t, "{", "}")

	err := &Error{
		Type:    "some error",
		Message: "bad list",
		Context: *contextInParen("[test]", "echo (x)"),
	}

	wantErrorString := "some error: [test]:1:6: bad list"
	if gotErrorString := err.Error(); gotErrorString != wantErrorString {
		t.Errorf("Error() -> %q, want %q", gotErrorString, wantErrorString)
	}

	wantRanging := Ranging{From: 5, To: 8}
	if gotRanging := err.Range(); gotRanging != wantRanging {
		t.Errorf("Range() -> %v, want %v", gotRanging, wantRanging)
	}

	// Type is capitalized in return value of Show
	wantShow := dedent(`
		Some error: {bad list}
		  [test]:1:6: echo <(x)>`)
	if gotShow := err.Show(""); gotShow != wantShow {
		t.Errorf("Show() -> %q, want %q", gotShow, wantShow)
	}
}
