package ui

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestParseSGREscapedText(t *testing.T) {
	tt.Test(t, tt.Fn("ParseSGREscapedText", ParseSGREscapedText), tt.Table{
		tt.Args("").Rets(Text(nil)),
		tt.Args("text").Rets(T("text")),
		tt.Args("\033[1mbold").Rets(T("bold", Bold)),
		tt.Args("\033[1mbold\033[31mbold red").Rets(
			Concat(T("bold", Bold), T("bold red", Bold, FgRed))),
		tt.Args("\033[1mbold\033[;31mred").Rets(
			Concat(T("bold", Bold), T("red", FgRed))),
		// Non-SGR CSI sequences are removed.
		tt.Args("\033[Atext").Rets(T("text")),
		// Control characters not part of CSI escape sequences are left
		// untouched.
		tt.Args("t\x01ext").Rets(T("t\x01ext")),
	})
}

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
