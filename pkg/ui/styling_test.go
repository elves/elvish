package ui

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestStyleText(t *testing.T) {
	tt.Test(t, tt.Fn("StyleText", StyleText), tt.Table{
		// Foreground color
		tt.Args(T("foo"), FgRed).
			Rets(Text{&Segment{Style{Foreground: Red}, "foo"}}),
		// Override existing foreground
		tt.Args(Text{&Segment{Style{Foreground: Green}, "foo"}}, FgRed).
			Rets(Text{&Segment{Style{Foreground: Red}, "foo"}}),
		// Multiple segments
		tt.Args(Text{
			&Segment{Style{}, "foo"},
			&Segment{Style{Foreground: Green}, "bar"}}, FgRed).
			Rets(Text{
				&Segment{Style{Foreground: Red}, "foo"},
				&Segment{Style{Foreground: Red}, "bar"},
			}),
		// Background color
		tt.Args(T("foo"), BgRed).
			Rets(Text{&Segment{Style{Background: Red}, "foo"}}),
		// Bold, false -> true
		tt.Args(T("foo"), Bold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// Bold, true -> true
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, Bold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// No Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, NoBold).
			Rets(Text{&Segment{Style{}, "foo"}}),
		// No Bold, false -> false
		tt.Args(T("foo"), NoBold).Rets(T("foo")),
		// Toggle Bold, true -> false
		tt.Args(Text{&Segment{Style{Bold: true}, "foo"}}, ToggleBold).
			Rets(Text{&Segment{Style{}, "foo"}}),
		// Toggle Bold, false -> true
		tt.Args(T("foo"), ToggleBold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// For the remaining bool transformers, we only check one case; the rest
		// should be similar to "bold".
		// Dim.
		tt.Args(T("foo"), Dim).
			Rets(Text{&Segment{Style{Dim: true}, "foo"}}),
		// Italic.
		tt.Args(T("foo"), Italic).
			Rets(Text{&Segment{Style{Italic: true}, "foo"}}),
		// Underlined.
		tt.Args(T("foo"), Underlined).
			Rets(Text{&Segment{Style{Underlined: true}, "foo"}}),
		// Blink.
		tt.Args(T("foo"), Blink).
			Rets(Text{&Segment{Style{Blink: true}, "foo"}}),
		// Inverse.
		tt.Args(T("foo"), Inverse).
			Rets(Text{&Segment{Style{Inverse: true}, "foo"}}),
		// TODO: Test nil styling.
	})
}

var parseStylingTests = []struct {
	s           string
	wantStyling Styling
}{
	{"default", FgDefault},
	{"red", FgRed},
	{"fg-default", FgDefault},
	{"fg-red", FgRed},

	{"bg-default", BgDefault},
	{"bg-red", BgRed},

	{"bold", Bold},
	{"no-bold", NoBold},
	{"toggle-bold", ToggleBold},

	{"red bold", Stylings(FgRed, Bold)},
}

func TestParseStyling(t *testing.T) {
	for _, test := range parseStylingTests {
		styling := ParseStyling(test.s)
		if !reflect.DeepEqual(styling, test.wantStyling) {
			t.Errorf("ParseStyling(%q) -> %v, want %v",
				test.s, styling, test.wantStyling)
		}
	}
}
