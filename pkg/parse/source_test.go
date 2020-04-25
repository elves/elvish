package parse_test

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	. "github.com/elves/elvish/pkg/parse"
	"github.com/xiaq/persistent/hash"
)

func TestSourceAsValue(t *testing.T) {
	vals.TestValue(t, &Source{Name: "[tty]", Code: "echo"}).
		Kind("map").
		Hash(hash.DJB(hash.String("[tty]"), hash.String("echo"), 0)).
		Equal(Source{Name: "[tty]", Code: "echo"}).
		NotEqual(Source{Name: "[tty]", Code: "put"}).
		Repr("<src name:'[tty]' code:... is-file:$false>").
		AllKeys("name", "code", "is-file").
		Index("name", "[tty]").
		Index("code", "echo").
		Index("is-file", false).
		Index("path", "")

	vals.TestValue(t, &Source{Name: "/etc/rc.elv", Code: "echo", IsFile: true}).
		Hash(hash.DJB(hash.String("/etc/rc.elv"), hash.String("echo"), 1)).
		Index("is-file", true).
		Index("path", "/etc/rc.elv")
}
