package parse_test

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	. "github.com/elves/elvish/pkg/parse"
)

func TestSourceAsStructMap(t *testing.T) {
	vals.TestValue(t, Source{Name: "[tty]", Code: "echo"}).
		Kind("structmap").
		Repr("[&name='[tty]' &code=<...> &is-file=$false]").
		AllKeys("name", "code", "is-file", "path").
		Index("path", "")

	vals.TestValue(t, Source{Name: "/etc/rc.elv", Code: "echo", IsFile: true}).
		Index("is-file", true).
		Index("path", "/etc/rc.elv")
}
