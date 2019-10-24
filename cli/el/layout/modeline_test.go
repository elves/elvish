package layout

import (
	"testing"

	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

func TestModeLine(t *testing.T) {
	testModeLine(t, tt.Fn("ModeLine", ModeLine))
}

func TestModePrompt(t *testing.T) {
	testModeLine(t, tt.Fn("ModePrompt",
		func(s string, b bool) styled.Text { return ModePrompt(s, b)() }))
}

func testModeLine(t *testing.T, fn *tt.FnToTest) {
	tt.Test(t, fn, tt.Table{
		tt.Args("TEST", false).Rets(
			styled.MakeText("TEST", "bold", "lightgray", "bg-magenta")),
		tt.Args("TEST", true).Rets(
			styled.MakeText("TEST", "bold", "lightgray", "bg-magenta").
				ConcatText(styled.Plain(" "))),
	})
}
