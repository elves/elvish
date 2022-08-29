package ui

import (
	"reflect"
	"testing"
)

func TestColorSGR(t *testing.T) {
	// Test the SGR sequences of colors indirectly via VTString of Text, since
	// that is how they are used.
	testTextVTString(t, []textVTStringTest{
		{T("foo", FgRed), "\033[;31mfoo\033[m"},
		{T("foo", BgRed), "\033[;41mfoo\033[m"},

		{T("foo", FgBrightRed), "\033[;91mfoo\033[m"},
		{T("foo", BgBrightRed), "\033[;101mfoo\033[m"},

		{T("foo", Fg(XTerm256Color(30))), "\033[;38;5;30mfoo\033[m"},
		{T("foo", Bg(XTerm256Color(30))), "\033[;48;5;30mfoo\033[m"},

		{T("foo", Fg(TrueColor(30, 40, 50))), "\033[;38;2;30;40;50mfoo\033[m"},
		{T("foo", Bg(TrueColor(30, 40, 50))), "\033[;48;2;30;40;50mfoo\033[m"},
	})
}

var colorStringTests = []struct {
	color Color
	str   string
}{
	{Red, "red"},
	{BrightRed, "bright-red"},
	{XTerm256Color(30), "color30"},
	{TrueColor(0x33, 0x44, 0x55), "#334455"},
}

func TestColorString(t *testing.T) {
	for _, test := range colorStringTests {
		s := test.color.String()
		if s != test.str {
			t.Errorf("%v.String() -> %q, want %q", test.color, s, test.str)
		}
	}
}

func TestParseColor(t *testing.T) {
	for _, test := range colorStringTests {
		c := parseColor(test.str)
		if !reflect.DeepEqual(c, test.color) {
			t.Errorf("parseError(%q) -> %v, want %v", test.str, c, test.color)
		}
	}
}
