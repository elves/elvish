package ui

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
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

	vals.TestValue(t, &Segment{Style{Foreground: Red, Background: Blue}, "foo"}).
		Repr("(ui:text-segment foo &fg-color=red &bg-color=blue)").
		Index("fg-color", "red").
		Index("bg-color", "blue")
}
