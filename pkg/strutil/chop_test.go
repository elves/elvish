package strutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestChopLineEnding(t *testing.T) {
	tt.Test(t, tt.Fn("ChopLineEnding", ChopLineEnding), tt.Table{
		tt.Args("").Rets(""),
		tt.Args("text").Rets("text"),
		tt.Args("text\n").Rets("text"),
		tt.Args("text\r\n").Rets("text"),
		// Only chop off one line ending
		tt.Args("text\n\n").Rets("text\n"),
		// Preserve internal line endings
		tt.Args("text\ntext 2\n").Rets("text\ntext 2"),
	})
}

func TestChopTerminator(t *testing.T) {
	tt.Test(t, tt.Fn("ChopTerminator", ChopTerminator), tt.Table{
		tt.Args("", byte('\x00')).Rets(""),
		tt.Args("foo", byte('\x00')).Rets("foo"),
		tt.Args("foo\x00", byte('\x00')).Rets("foo"),
		tt.Args("foo\x00\x00", byte('\x00')).Rets("foo\x00"),
		tt.Args("foo\x00bar\x00", byte('\x00')).Rets("foo\x00bar"),
	})
}
