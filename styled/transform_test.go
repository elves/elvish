package styled

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestTransform(t *testing.T) {
	tt.Test(t, tt.Fn("Transform", Transform), tt.Table{
		// Foreground color
		tt.Args(Unstyled("foo"), "red").
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
		tt.Args(Unstyled("foo"), "bg-red").
			Rets(Text{&Segment{Style{Background: "red"}, "foo"}}),
		// Bold
		tt.Args(Unstyled("foo"), "bold").
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// No Bold
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, "no-bold").
			Rets(Text{&Segment{Style{}, "foo"}}),
		// Toggle Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, "toggle-bold").
			Rets(Text{&Segment{Style{}, "foo"}}),
		// Toggle Bold, false -> true
		tt.Args(Unstyled("foo"), "toggle-bold").
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// Invalid transformer
		tt.Args(Unstyled("foo"), "invalid").
			Rets(Text{&Segment{Text: "foo"}}),
	})
}
