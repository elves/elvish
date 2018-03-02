package completion

import (
	"fmt"
	"unicode"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/util"
)

func throw(e error) {
	util.Throw(e)
}

func maybeThrow(e error) {
	if e != nil {
		util.Throw(e)
	}
}

func throwf(format string, args ...interface{}) {
	util.Throw(fmt.Errorf(format, args...))
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

// Styles for UI.
var (
	// Use default style for completion listing
	styleForCompletion = ui.Styles{}
	// Use inverse style for selected completion entry
	styleForSelectedCompletion = ui.Styles{"inverse"}
)

// likeChar returns if a key looks like a character meant to be input (as
// opposed to a function key).
func likeChar(k ui.Key) bool {
	return k.Mod == 0 && k.Rune > 0 && unicode.IsGraphic(k.Rune)
}
