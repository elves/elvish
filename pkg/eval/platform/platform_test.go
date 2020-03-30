package platform

import (
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestPlatform(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("platform", Ns) }
	eval.TestWithSetup(t, setup,
		That(`put $platform:arch`).Puts(runtime.GOARCH),
		That(`put $platform:os`).Puts(runtime.GOOS),
		That(`put $platform:is-windows`).Puts(runtime.GOOS == "windows"),
		That(`put $platform:is-unix`).Puts(
			runtime.GOOS != "windows" && runtime.GOOS != "plan9" && runtime.GOOS != "js"),
	)
}
