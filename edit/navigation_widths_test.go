package edit

import (
	"testing"

	"github.com/elves/elvish/tt"
)

var (
	getNavWidthsTests = tt.Table{
		// Enough room for both current and preview: parent gets 1/6, current
		// and preview gets 1/2 of remain
		tt.Args(120, 10, 10).Rets(20, 50, 50),
		// Not enough room for either of current and preview: same as above
		tt.Args(120, 100, 100).Rets(20, 50, 50),
		// Enough room for current but not preview; current donates to preview
		tt.Args(120, 10, 100).Rets(20, 10, 90),
		// Enough room for preview but not current; preview donates to current
		tt.Args(120, 100, 10).Rets(20, 90, 10),
	}
)

func TestGetNavWidths(t *testing.T) {
	tt.Test(t, tt.Fn("getNavWidths", getNavWidths), getNavWidthsTests)
}
