package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/xiaq/persistent/hash"
)

func TestSourceAsValue(t *testing.T) {
	vals.TestValue(t, NewInteractiveSource("[tty]", "echo")).
		Kind("map").
		Hash(hash.DJB(uint32(InteractiveSource),
			hash.String("[tty]"), 1, hash.String("echo"))).
		Equal(NewInteractiveSource("[tty]", "echo")).
		NotEqual(NewInteractiveSource("[tty]", "put")).
		Repr("<src type:interactive name:'[tty]' root:$true code:...>").
		AllKeys("type", "name", "root", "code").
		Index("type", "interactive").
		Index("name", "[tty]").
		Index("root", true).
		Index("code", "echo")

	vals.TestValue(t, NewInternalGoSource("[test]")).
		Index("type", "internal-go")
	vals.TestValue(t, NewInternalElvishSource(true, "[test]", "echo")).
		Index("type", "internal-elvish")
	vals.TestValue(t, NewScriptSource("/fake/path", "echo")).
		Index("type", "file")
	vals.TestValue(t, &Source{Type: -1}).
		Index("type", "bad type -1")
}
