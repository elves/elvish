package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestStyleFromSGR(t *testing.T) {
	tt.Test(t, tt.Fn("StyleFromSGR", StyleFromSGR), tt.Table{
		tt.Args("1").Rets(Style{Bold: true}),
		// Multiple codes
		tt.Args("31;42").Rets(Style{Foreground: "red", Background: "green"}),
		// Invalid codes are ignored
		tt.Args("1;invalid;10000").Rets(Style{Bold: true}),
	})
}
