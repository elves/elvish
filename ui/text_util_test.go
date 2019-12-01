package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMarkLines(t *testing.T) {
	stylesheet := map[rune]Styling{
		'-': Inverse,
		'x': JoinStylings(Blue, BgGreen),
	}
	tt.Test(t, tt.Fn("MarkLines", MarkLines), tt.Table{
		tt.Args("foo  bar foobar").Rets(T("foo  bar foobar")),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
		).Rets(
			T("foo", Inverse).
				ConcatText(T("  ")).
				ConcatText(T("bar", Blue, BgGreen)).
				ConcatText(T(" ")).
				ConcatText(T("foobar", Inverse)),
		),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			T("foo", Inverse).
				ConcatText(T("  bar foobar")),
		),
		tt.Args(
			"plain1",
			"plain2",
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
			"plain3",
		).Rets(
			T("plain1").
				ConcatText(T("\n")).
				ConcatText(T("plain2")).
				ConcatText(T("\n")).
				ConcatText(T("foo", Inverse)).
				ConcatText(T("  ")).
				ConcatText(T("bar", Blue, BgGreen)).
				ConcatText(T(" ")).
				ConcatText(T("foobar", Inverse)).
				ConcatText(T("\n")).
				ConcatText(T("plain3")),
		),
	})
}
