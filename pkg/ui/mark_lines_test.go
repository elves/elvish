package ui

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestMarkLines(t *testing.T) {
	stylesheet := RuneStylesheet{
		'-': Inverse,
		'x': Stylings(FgBlue, BgGreen),
	}
	tt.Test(t, MarkLines,
		Args("foo  bar foobar").Rets(T("foo  bar foobar")),
		Args(
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
		Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			Concat(
				T("foo", Inverse),
				T("  bar foobar")),
		),
		Args(
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
	)
}
