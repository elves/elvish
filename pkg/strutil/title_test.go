package strutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestTitle(t *testing.T) {
	tt.Test(t, Title,
		Args("").Rets(""),
		Args("foo").Rets("Foo"),
		Args("\xf0").Rets("\xf0"),
		Args("FOO").Rets("FOO"),
	)
}
