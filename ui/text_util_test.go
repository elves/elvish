package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestMarkLines(t *testing.T) {
	stylesheet := map[rune]string{
		'-': "reverse",
		'x': "blue bg-green",
	}
	tt.Test(t, tt.Fn("MarkLines", MarkLines), tt.Table{
		tt.Args("foo  bar foobar").Rets(PlainText("foo  bar foobar")),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
		).Rets(
			MakeText("foo", "reverse").
				ConcatText(PlainText("  ")).
				ConcatText(MakeText("bar", "blue", "bg-green")).
				ConcatText(PlainText(" ")).
				ConcatText(MakeText("foobar", "reverse")),
		),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			MakeText("foo", "reverse").
				ConcatText(PlainText("  bar foobar")),
		),
		tt.Args(
			"plain1",
			"plain2",
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
			"plain3",
		).Rets(
			PlainText("plain1").
				ConcatText(PlainText("\n")).
				ConcatText(PlainText("plain2")).
				ConcatText(PlainText("\n")).
				ConcatText(MakeText("foo", "reverse")).
				ConcatText(PlainText("  ")).
				ConcatText(MakeText("bar", "blue", "bg-green")).
				ConcatText(PlainText(" ")).
				ConcatText(MakeText("foobar", "reverse")).
				ConcatText(PlainText("\n")).
				ConcatText(PlainText("plain3")),
		),
	})
}
