package ui

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestStyleFromSGR(t *testing.T) {
	tt.Test(t, tt.Fn("StyleFromSGR", StyleFromSGR), tt.Table{
		tt.Args("1").Rets(Style{Bold: true}),
		// Invalid codes are ignored
		tt.Args("1;invalid;10000").Rets(Style{Bold: true}),
		// ANSI colors.
		tt.Args("31;42").Rets(Style{Foreground: Red, Background: Green}),
		// ANSI bright colors.
		tt.Args("91;102").
			Rets(Style{Foreground: BrightRed, Background: BrightGreen}),
		// XTerm 256 color.
		tt.Args("38;5;1;48;5;2").
			Rets(Style{Foreground: XTerm256Color(1), Background: XTerm256Color(2)}),
		// True colors.
		tt.Args("38;2;1;2;3;48;2;10;20;30").
			Rets(Style{
				Foreground: TrueColor(1, 2, 3), Background: TrueColor(10, 20, 30)}),
	})
}
