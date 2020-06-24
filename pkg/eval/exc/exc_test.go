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

		That("exc:is-external-cmd-exc ?("+failingExternalCmd+")").Puts(true),
		That("exc:is-external-cmd-exc ?(fail bad)").Puts(false),

		That("exc:is-nonzero-exit ?("+failingExternalCmd+")").Puts(true),
		That("exc:is-nonzero-exit ?(fail bad)").Puts(false),

		// TODO: Test positive case of exc:is-killed
		That("exc:is-killed ?(fail bad)").Puts(false),

		That("exc:is-fail-exc ?(fail bad)").Puts(true),
		That("exc:is-fail-exc ?("+failingExternalCmd+")").Puts(false),

		That("exc:is-pipeline-exc ?(fail bad)").Puts(false),
		That("exc:is-pipeline-exc ?(fail 1 | fail 2)").Puts(true),
	)
}
