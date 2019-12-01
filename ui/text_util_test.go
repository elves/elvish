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
		tt.Args("foo  bar foobar").Rets(MakeText("foo  bar foobar")),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
		).Rets(
			MakeText("foo", "reverse").
				ConcatText(MakeText("  ")).
				ConcatText(MakeText("bar", "blue", "bg-green")).
				ConcatText(MakeText(" ")).
				ConcatText(MakeText("foobar", "reverse")),
		),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			MakeText("foo", "reverse").
				ConcatText(MakeText("  bar foobar")),
		),
		tt.Args(
			"plain1",
			"plain2",
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
			"plain3",
		).Rets(
			MakeText("plain1").
				ConcatText(MakeText("\n")).
				ConcatText(MakeText("plain2")).
				ConcatText(MakeText("\n")).
				ConcatText(MakeText("foo", "reverse")).
				ConcatText(MakeText("  ")).
				ConcatText(MakeText("bar", "blue", "bg-green")).
				ConcatText(MakeText(" ")).
				ConcatText(MakeText("foobar", "reverse")).
				ConcatText(MakeText("\n")).
				ConcatText(MakeText("plain3")),
		),
	})
}
