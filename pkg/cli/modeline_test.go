package cli

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
	"github.com/elves/elvish/pkg/ui"
)

func TestModeLine(t *testing.T) {
	testModeLine(t, tt.Fn("ModeLine", ModeLine))
}

func TestModePrompt(t *testing.T) {
	testModeLine(t, tt.Fn("ModePrompt",
		func(s string, b bool) ui.Text { return ModePrompt(s, b)() }))
}

func testModeLine(t *testing.T, fn *tt.FnToTest) {
	tt.Test(t, fn, tt.Table{
		tt.Args("TEST", false).Rets(
			ui.T("TEST", ui.Bold, ui.FgWhite, ui.BgMagenta)),
		tt.Args("TEST", true).Rets(
			ui.T("TEST", ui.Bold, ui.FgWhite, ui.BgMagenta).
				ConcatText(ui.T(" "))),
	})
}
