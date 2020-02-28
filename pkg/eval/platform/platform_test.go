package platform

import (
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestOs(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("platform", Ns) }
	eval.TestWithSetup(t, setup,
		That(`put $platform:arch`).Puts(runtime.GOARCH),
		That(`put $platform:os`).Puts(runtime.GOOS),
	)
}
