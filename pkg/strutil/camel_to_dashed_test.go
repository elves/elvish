package strutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestCamelToDashed(t *testing.T) {
	tt.Test(t, tt.Fn("CamelToDashed", CamelToDashed), tt.Table{
		Args("CamelCase").Rets("camel-case"),
		Args("camelCase").Rets("-camel-case"),
		Args("HTTP").Rets("http"),
		Args("HTTPRequest").Rets("http-request"),
	})
}
