package lscolors

import (
	"os"
	"testing"

	"github.com/elves/elvish/util"
)

func TestLsColors(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()
	restoreLsColors := WithTestLsColors()
	defer restoreLsColors()

	// Test both feature-based and extension-based coloring.

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
