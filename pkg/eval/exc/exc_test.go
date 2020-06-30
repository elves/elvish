package exc

import (
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestExc(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Global.AddNs("exc", Ns) }
	eval.TestWithSetup(t, setup,
		// Have a simple sanity test that exc:show writes something.
		That(`exc:show ?(fail foo) | > (count (slurp)) 0`).Puts(true),
	)
}
