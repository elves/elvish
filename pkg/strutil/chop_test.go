package strutil

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

func TestChopLineEnding(t *testing.T) {
	Test(t, Fn("ChopLineEnding", ChopLineEnding), Table{
		Args("").Rets(""),
		Args("text").Rets("text"),
		Args("text\n").Rets("text"),
		Args("text\r\n").Rets("text"),
		// Only chop off one line ending
		Args("text\n\n").Rets("text\n"),
		// Preserve internal line endings
		Args("text\ntext 2\n").Rets("text\ntext 2"),
	})
}
