package styled

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
		tt.Args("foo  bar foobar").Rets(Plain("foo  bar foobar")),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
		).Rets(
			MakeText("foo", "reverse").
				ConcatText(Plain("  ")).
				ConcatText(MakeText("bar", "blue", "bg-green")).
				ConcatText(Plain(" ")).
				ConcatText(MakeText("foobar", "reverse")),
		),
		tt.Args(
			"foo  bar foobar", stylesheet,
			"---",
		).Rets(
			MakeText("foo", "reverse").
				ConcatText(Plain("  bar foobar")),
		),
		tt.Args(
			"plain1",
			"plain2",
			"foo  bar foobar", stylesheet,
			"---  xxx ------",
			"plain3",
		).Rets(
			Plain("plain1").
				ConcatText(Plain("\n")).
				ConcatText(Plain("plain2")).
				ConcatText(Plain("\n")).
				ConcatText(MakeText("foo", "reverse")).
				ConcatText(Plain("  ")).
				ConcatText(MakeText("bar", "blue", "bg-green")).
				ConcatText(Plain(" ")).
				ConcatText(MakeText("foobar", "reverse")).
				ConcatText(Plain("\n")).
				ConcatText(Plain("plain3")),
		),
	})
}
