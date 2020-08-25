package ui

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestMarkLines(t *testing.T) {
	stylesheet := RuneStylesheet{
		'-': Inverse,
		'x': Stylings(FgBlue, BgGreen),
	}
	tt.Test(t, tt.Fn("MarkLines", MarkLines), tt.Table{
		tt.Args("foo  bar foobar").Rets(T("foo  bar foobar")),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
		).Rets(
			Concat(
				T("foo", Inverse),
				T("  "),
				T("bar", FgBlue, BgGreen),
				T(" "),
				T("foobar", Inverse)),
		),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			Concat(
				T("foo", Inverse),
				T("  bar foobar")),
		),
		tt.Args(
			"plain1",
			"plain2",
			"foo  bar foobar\n", stylesheet,
			"---  xxx ------",
			"plain3",
		).Rets(
			Concat(
				T("plain1"),
				T("plain2"),
				T("foo", Inverse),
				T("  "),
				T("bar", FgBlue, BgGreen),
				T(" "),
				T("foobar", Inverse),
				T("\n"),
				T("plain3")),
		),
	})
}
