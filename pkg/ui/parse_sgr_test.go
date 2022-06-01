package ui

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestParseSGREscapedText(t *testing.T) {
	tt.Test(t, tt.Fn("ParseSGREscapedText", ParseSGREscapedText), tt.Table{
		Args("").Rets(Text(nil)),
		Args("text").Rets(T("text")),
		Args("\033[1mbold").Rets(T("bold", Bold)),
		Args("\033[1mbold\033[31mbold red").Rets(
			Concat(T("bold", Bold), T("bold red", Bold, FgRed))),
		Args("\033[1mbold\033[;31mred").Rets(
			Concat(T("bold", Bold), T("red", FgRed))),
		// Non-SGR CSI sequences are removed.
		Args("\033[Atext").Rets(T("text")),
		// Control characters not part of CSI escape sequences are left
		// untouched.
		Args("t\x01ext").Rets(T("t\x01ext")),
	})
}

func TestStyleFromSGR(t *testing.T) {
	tt.Test(t, tt.Fn("StyleFromSGR", StyleFromSGR), tt.Table{
		Args("1").Rets(Style{Bold: true}),
		// Invalid codes are ignored
		Args("1;invalid;10000").Rets(Style{Bold: true}),
		// ANSI colors.
		Args("31;42").Rets(Style{Foreground: Red, Background: Green}),
		// ANSI bright colors.
		Args("91;102").
			Rets(Style{Foreground: BrightRed, Background: BrightGreen}),
		// XTerm 256 color.
		Args("38;5;1;48;5;2").
			Rets(Style{Foreground: XTerm256Color(1), Background: XTerm256Color(2)}),
		// True colors.
		Args("38;2;1;2;3;48;2;10;20;30").
			Rets(Style{
				Foreground: TrueColor(1, 2, 3), Background: TrueColor(10, 20, 30)}),
	})
}
