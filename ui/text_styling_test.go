package ui

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestStyleText(t *testing.T) {
	tt.Test(t, tt.Fn("StyleText", StyleText), tt.Table{
		// Foreground color
		tt.Args(NewText("foo"), Red).
			Rets(Text{&Segment{Style{Foreground: "red"}, "foo"}}),
		// Override existing foreground
		tt.Args(Text{&Segment{Style{Foreground: "green"}, "foo"}}, Red).
			Rets(Text{&Segment{Style{Foreground: "red"}, "foo"}}),
		// Multiple segments
		tt.Args(Text{
			&Segment{Style{}, "foo"},
			&Segment{Style{Foreground: "green"}, "bar"}}, Red).
			Rets(Text{
				&Segment{Style{Foreground: "red"}, "foo"},
				&Segment{Style{Foreground: "red"}, "bar"},
			}),
		// Background color
		tt.Args(NewText("foo"), RedBackground).
			Rets(Text{&Segment{Style{Background: "red"}, "foo"}}),
		// Bold, false -> true
		tt.Args(NewText("foo"), Bold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// Bold, true -> true
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, Bold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// No Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, NoBold).
			Rets(Text{&Segment{Style{}, "foo"}}),
		// No Bold, false -> false
		tt.Args(NewText("foo"), NoBold).Rets(NewText("foo")),
		// Toggle Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, ToggleBold).
			Rets(Text{&Segment{Style{}, "foo"}}),
		// Toggle Bold, false -> true
		tt.Args(NewText("foo"), ToggleBold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// For the remaining bool transformers, we only check one case; the rest
		// should be similar to "bold".
		// Dim.
		tt.Args(NewText("foo"), Dim).
			Rets(Text{&Segment{Style{Dim: true}, "foo"}}),
		// Italic.
		tt.Args(NewText("foo"), Italic).
			Rets(Text{&Segment{Style{Italic: true}, "foo"}}),
		// Underlined.
		tt.Args(NewText("foo"), Underlined).
			Rets(Text{&Segment{Style{Underlined: true}, "foo"}}),
		// Blink.
		tt.Args(NewText("foo"), Blink).
			Rets(Text{&Segment{Style{Blink: true}, "foo"}}),
		// Inverse.
		tt.Args(NewText("foo"), Inverse).
			Rets(Text{&Segment{Style{Inverse: true}, "foo"}}),
		// TODO: Test nil styling.
	})
}
