package ui

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestStyleText(t *testing.T) {
	tt.Test(t, StyleText,
		// Foreground color
		Args(T("foo"), FgRed).
			Rets(Text{&Segment{Style{Fg: Red}, "foo"}}),
		// Override existing foreground
		Args(Text{&Segment{Style{Fg: Green}, "foo"}}, FgRed).
			Rets(Text{&Segment{Style{Fg: Red}, "foo"}}),
		// Multiple segments
		Args(Text{
			&Segment{Style{}, "foo"},
			&Segment{Style{Fg: Green}, "bar"}}, FgRed).
			Rets(Text{
				&Segment{Style{Fg: Red}, "foo"},
				&Segment{Style{Fg: Red}, "bar"},
			}),
		// Background color
		Args(T("foo"), BgRed).
			Rets(Text{&Segment{Style{Bg: Red}, "foo"}}),
		// Bold, false -> true
		Args(T("foo"), Bold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// Bold, true -> true
		Args(Text{&Segment{Style{Bold: true}, "foo"}}, Bold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// No Bold, true -> false
		Args(Text{&Segment{Style{Bold: true}, "foo"}}, NoBold).
			Rets(Text{&Segment{Style{}, "foo"}}),
		// No Bold, false -> false
		Args(T("foo"), NoBold).Rets(T("foo")),
		// Toggle Bold, true -> false
		Args(Text{&Segment{Style{Bold: true}, "foo"}}, ToggleBold).
			Rets(Text{&Segment{Style{}, "foo"}}),
		// Toggle Bold, false -> true
		Args(T("foo"), ToggleBold).
			Rets(Text{&Segment{Style{Bold: true}, "foo"}}),
		// For the remaining bool transformers, we only check one case; the rest
		// should be similar to "bold".
		// Dim.
		Args(T("foo"), Dim).
			Rets(Text{&Segment{Style{Dim: true}, "foo"}}),
		// Italic.
		Args(T("foo"), Italic).
			Rets(Text{&Segment{Style{Italic: true}, "foo"}}),
		// Underlined.
		Args(T("foo"), Underlined).
			Rets(Text{&Segment{Style{Underlined: true}, "foo"}}),
		// Blink.
		Args(T("foo"), Blink).
			Rets(Text{&Segment{Style{Blink: true}, "foo"}}),
		// Inverse.
		Args(T("foo"), Inverse).
			Rets(Text{&Segment{Style{Inverse: true}, "foo"}}),
		// TODO: Test nil styling.
	)
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
