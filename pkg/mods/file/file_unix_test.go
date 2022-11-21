//go:build !windows && !plan9 && !js

package file

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
)

func TestIsTTY(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("file", Ns))
	}

	evaltest.TestWithSetup(t, setup,
		That("file:is-tty 0 < /dev/null").Puts(false),
		That("file:is-tty (num 0) < /dev/null").Puts(false),
		That("file:is-tty 0 < /dev/tty").Puts(true),
		That("file:is-tty (num 0) < /dev/tty").Puts(true),
	)
}
