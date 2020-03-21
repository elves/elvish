// +build windows plan9 js

package unix

import (
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestUnix(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("unix", Ns) }
	eval.TestWithSetup(t, setup,
		That(`== 0 (count [(keys $unix:)])`).Puts(true),
	)
}
