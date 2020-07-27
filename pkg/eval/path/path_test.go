package path

import (
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestStr(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("path", Ns) }
	eval.TestWithSetup(t, setup,
		That(`path:base a/b/c.png`).Puts("c.png"),
		That(`path:ext a/b/c.png`).Puts(".png"),
	)
}
