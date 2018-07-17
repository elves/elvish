package styled

import (
	"testing"

	"github.com/elves/elvish/tt"
)

func TestTransform(t *testing.T) {
	tt.Test(t, tt.Fn("Transform", Transform), tt.Table{
		// Foreground color
		tt.Args(Text{&Segment{Style{}, "text"}}, "red").Rets(
			Text{&Segment{Text: "text", Style: Style{Foreground: "red"}}}),
		// Override existing foreground
		tt.Args(Text{&Segment{Style{Foreground: "green"}, "text"}}, "red").Rets(
			Text{&Segment{Text: "text", Style: Style{Foreground: "red"}}}),
		// Background color
		tt.Args(Text{&Segment{Style{}, "text"}}, "bg-red").Rets(
			Text{&Segment{Text: "text", Style: Style{Background: "red"}}}),
		// Invalid transformer
		tt.Args(Text{&Segment{Style{}, "text"}}, "invalid").Rets(
			Text{&Segment{Text: "text"}}),
	})
}
