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
		Equal(&Source{Name: "[tty]", Code: "echo"}).
		NotEqual(&Source{Name: "[tty]", Code: "put"}).
		Repr("<src name:'[tty]' code:... is-file:$false>").
		AllKeys("name", "code", "is-file").
		Index("name", "[tty]").
		Index("code", "echo").
		Index("is-file", false)

	vals.TestValue(t, &Source{IsFile: true}).
		Index("is-file", true)
}
