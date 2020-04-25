package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/xiaq/persistent/hash"
)

func TestSourceAsValue(t *testing.T) {
	vals.TestValue(t, NewInteractiveSource("[tty]", "echo")).
		Kind("map").
		Hash(hash.DJB(hash.String("[tty]"), hash.String("echo"), 0)).
		Equal(NewInteractiveSource("[tty]", "echo")).
		NotEqual(NewInteractiveSource("[tty]", "put")).
		Repr("<src name:'[tty]' code:... is-file:$false>").
		AllKeys("name", "code", "is-file").
		Index("name", "[tty]").
		Index("code", "echo").
		Index("is-file", false)

	vals.TestValue(t, NewInternalGoSource("[test]")).
		Index("is-file", false)

	vals.TestValue(t, NewInternalElvishSource(true, "[test]", "echo")).
		Index("is-file", false)

	vals.TestValue(t, NewScriptSource("/fake/path", "echo")).
		Index("is-file", true)
}
