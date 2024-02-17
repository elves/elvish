package strutil_test

import (
	"testing"

	"src.elv.sh/pkg/strutil"
	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestJoinLines(t *testing.T) {
	tt.Test(t, strutil.JoinLines,
		Args([]string(nil)).Rets(""),
		Args([]string{"foo"}).Rets("foo\n"),
		Args([]string{"foo", "bar"}).Rets("foo\nbar\n"),
	)
}
