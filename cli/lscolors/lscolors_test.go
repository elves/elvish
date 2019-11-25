package lscolors

import (
	"os"
	"testing"

	"github.com/elves/elvish/util"
)

func TestLsColors(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()
	savedLsColors := os.Getenv("LS_COLORS")
	defer os.Setenv("LS_COLORS", savedLsColors)

	// Test both feature-based and extension-based coloring. Directory is blue,
	// .png files are red.
	os.Setenv("LS_COLORS", "di=34:*.png=31")
	colorist := GetColorist()

	os.Mkdir("dir", 0755)
	create("a.png", 0644)

	wantDirStyle := "34"
	if style := colorist.GetStyle("dir"); style != wantDirStyle {
		t.Errorf("Got dir style %q, want %q", style, wantDirStyle)
	}
	wantPngStyle := "31"
	if style := colorist.GetStyle("a.png"); style != wantPngStyle {
		t.Errorf("Got dir style %q, want %q", style, wantPngStyle)
	}
}
