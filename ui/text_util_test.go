package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMarkLines(t *testing.T) {
	stylesheet := map[rune]Transformer{
		'-': Inverse,
		'x': JoinTransformers(Blue, GreenBackground),
	}
	tt.Test(t, tt.Fn("MarkLines", MarkLines), tt.Table{
		tt.Args("foo  bar foobar").Rets(NewText("foo  bar foobar")),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
		).Rets(
			NewText("foo", Inverse).
				ConcatText(NewText("  ")).
				ConcatText(NewText("bar", Blue, GreenBackground)).
				ConcatText(NewText(" ")).
				ConcatText(NewText("foobar", Inverse)),
		),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			NewText("foo", Inverse).
				ConcatText(NewText("  bar foobar")),
		),
		tt.Args(
			"plain1",
			"plain2",
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
			"plain3",
		).Rets(
			NewText("plain1").
				ConcatText(NewText("\n")).
				ConcatText(NewText("plain2")).
				ConcatText(NewText("\n")).
				ConcatText(NewText("foo", Inverse)).
				ConcatText(NewText("  ")).
				ConcatText(NewText("bar", Blue, GreenBackground)).
				ConcatText(NewText(" ")).
				ConcatText(NewText("foobar", Inverse)).
				ConcatText(NewText("\n")).
				ConcatText(NewText("plain3")),
		),
	})
}
