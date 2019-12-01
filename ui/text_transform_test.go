package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestTransformText(t *testing.T) {
	tt.Test(t, tt.Fn("TransformText", TransformText), tt.Table{
		// Foreground color
		tt.Args(MakeText("foo"), "red").
			Rets(Text{&Segment{Style{Foreground: "red"}, "foo"}}),
		// Override existing foreground
		tt.Args(Text{&Segment{Style{Foreground: "green"}, "foo"}}, "red").
			Rets(Text{&Segment{Style{Foreground: "red"}, "foo"}}),
		// Multiple segments
		tt.Args(Text{
			&Segment{Style{}, "foo"},
			&Segment{Style{Foreground: "green"}, "bar"}}, "red").
			Rets(Text{
				&Segment{Style{Foreground: "red"}, "foo"},
				&Segment{Style{Foreground: "red"}, "bar"},
			}),
		// Background color
		tt.Args(MakeText("foo"), "bg-red").
			Rets(Text{&Segment{Style{Background: "red"}, "foo"}}),
		// Bold, false -> true
		tt.Args(MakeText("foo"), "bold").
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// Bold, true -> true
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, "bold").
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// No Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, "no-bold").
			Rets(Text{&Segment{Style{}, "foo"}}),
		// No Bold, false -> false
		tt.Args(MakeText("foo"), "no-bold").Rets(MakeText("foo")),
		// Toggle Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, "toggle-bold").
			Rets(Text{&Segment{Style{}, "foo"}}),
		// Toggle Bold, false -> true
		tt.Args(MakeText("foo"), "toggle-bold").
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// For the remaining bool transformers, we only check one case; the rest
		// should be similar to "bold".
		// Dim.
		tt.Args(MakeText("foo"), "dim").
			Rets(Text{&Segment{Style{Dim: true}, "foo"}}),
		// Italic.
		tt.Args(MakeText("foo"), "italic").
			Rets(Text{&Segment{Style{Italic: true}, "foo"}}),
		// Underlined.
		tt.Args(MakeText("foo"), "underlined").
			Rets(Text{&Segment{Style{Underlined: true}, "foo"}}),
		// Blink.
		tt.Args(MakeText("foo"), "blink").
			Rets(Text{&Segment{Style{Blink: true}, "foo"}}),
		// Inverse.
		tt.Args(MakeText("foo"), "inverse").
			Rets(Text{&Segment{Style{Inverse: true}, "foo"}}),
		// Invalid transformer
		tt.Args(MakeText("foo"), "invalid").
			Rets(Text{&Segment{Text: "foo"}}),
	})
}
