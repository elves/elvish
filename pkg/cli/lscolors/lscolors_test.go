package lscolors

import (
	"os"
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestLsColors(t *testing.T) {
	SetTestLsColors(t)
	testutil.InTempDir(t)
	os.Mkdir("dir", 0755)
	create("a.png")

	colorist := GetColorist()

	// Feature-based coloring.
	wantDirStyle := "34"
	if style := colorist.GetStyle("dir"); style != wantDirStyle {
		t.Errorf("Got dir style %q, want %q", style, wantDirStyle)
	}
	// Extension-based coloring.
	wantPngStyle := "31"
	if style := colorist.GetStyle("a.png"); style != wantPngStyle {
		t.Errorf("Got dir style %q, want %q", style, wantPngStyle)
	}
}

func TestLsColors_SkipsInvalidFields(t *testing.T) {
	testutil.Setenv(t, "LS_COLORS", "invalid=34:*.png=31")
	testutil.InTempDir(t)
	create("a.png")

	wantPngStyle := "31"
	if style := GetColorist().GetStyle("a.png"); style != wantPngStyle {
		t.Errorf("Got dir style %q, want %q", style, wantPngStyle)
	}
}

func TestLsColors_Default(t *testing.T) {
	testutil.Setenv(t, "LS_COLORS", "")
	testutil.InTempDir(t)
	create("a.png")

	// See defaultLsColorString
	wantPngStyle := "01;35"
	if style := GetColorist().GetStyle("a.png"); style != wantPngStyle {
		t.Errorf("Got dir style %q, want %q", style, wantPngStyle)
	}
}
