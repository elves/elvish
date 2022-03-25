package strutil

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestCamelToDashed(t *testing.T) {
	tt.Test(t, tt.Fn("CamelToDashed", CamelToDashed), tt.Table{
		tt.Args("CamelCase").Rets("camel-case"),
		tt.Args("camelCase").Rets("-camel-case"),
		tt.Args("HTTP").Rets("http"),
		tt.Args("HTTPRequest").Rets("http-request"),
	})
}
