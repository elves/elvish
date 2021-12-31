package strutil

import (
	"testing"

	. "src.elv.sh/pkg/tt"
)

func TestCamelToDashed(t *testing.T) {
	Test(t, Fn("CamelToDashed", CamelToDashed), Table{
		Args("CamelCase").Rets("camel-case"),
		Args("camelCase").Rets("-camel-case"),
		Args("HTTP").Rets("http"),
		Args("HTTPRequest").Rets("http-request"),
	})
}
