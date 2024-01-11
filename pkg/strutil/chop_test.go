package strutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestChopLineEnding(t *testing.T) {
	tt.Test(t, ChopLineEnding,
		Args("").Rets(""),
		Args("text").Rets("text"),
		Args("text\n").Rets("text"),
		Args("text\r\n").Rets("text"),
		// Only chop off one line ending
		Args("text\n\n").Rets("text\n"),
		// Preserve internal line endings
		Args("text\ntext 2\n").Rets("text\ntext 2"),
	)
}

func TestChopTerminator(t *testing.T) {
	tt.Test(t, ChopTerminator,
		Args("", byte('\x00')).Rets(""),
		Args("foo", byte('\x00')).Rets("foo"),
		Args("foo\x00", byte('\x00')).Rets("foo"),
		Args("foo\x00\x00", byte('\x00')).Rets("foo\x00"),
		Args("foo\x00bar\x00", byte('\x00')).Rets("foo\x00bar"),
	)
}
