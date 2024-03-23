package ui

import (
	"testing"

	"src.elv.sh/pkg/eval/vals"
)

func TestTextSegmentAsElvishValue(t *testing.T) {
	vals.TestValue(t, &Segment{Style{}, "foo"}).
		Kind("ui:text-segment").
		Repr("foo").
		AllKeys("text", "fg-color", "bg-color",
			"bold", "dim", "italic", "underlined", "blink", "inverse").
		Index("text", "foo").
		Index("fg-color", "default").
		Index("bg-color", "default").
		Index("bold", false).
		Index("dim", false).
		Index("italic", false).
		Index("underlined", false).
		Index("blink", false).
		Index("inverse", false)

	vals.TestValue(t, &Segment{Style{Fg: Red, Bg: Blue}, "foo"}).
		Repr("(styled-segment foo &fg-color=red &bg-color=blue)").
		Index("fg-color", "red").
		Index("bg-color", "blue")
}

var textSegmentVTStringTests = []struct {
	name         string
	seg          *Segment
	wantVTString string
}{
	{
		name:         "seg with no style",
		seg:          &Segment{Text: "foo"},
		wantVTString: "\033[mfoo",
	},
	{
		name:         "seg with style",
		seg:          &Segment{Style: Style{Bold: true}, Text: "foo"},
		wantVTString: "\033[;1mfoo\033[m",
	},
}

func TestTextSegmentVTString(t *testing.T) {
	for _, tc := range textSegmentVTStringTests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.seg.VTString(); got != tc.wantVTString {
				t.Errorf("VTString of %#v is %q, want %q", tc.seg, got, tc.wantVTString)
			}
			if got := tc.seg.String(); got != tc.wantVTString {
				t.Errorf("String of %#v is %q, want %q", tc.seg, got, tc.wantVTString)
			}
		})
	}
}
