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

		That("exc:is-external-cmd-err ?("+failingExternalCmd+")[reason]").Puts(true),
		That("exc:is-external-cmd-err ?(fail bad)[reason]").Puts(false),

		That("exc:is-nonzero-exit ?("+failingExternalCmd+")[reason]").Puts(true),
		That("exc:is-nonzero-exit ?(fail bad)[reason]").Puts(false),

		// TODO: Test positive case of exc:is-killed
		That("exc:is-killed ?(fail bad)[reason]").Puts(false),

		That("exc:is-fail-err ?(fail bad)[reason]").Puts(true),
		That("exc:is-fail-err ?("+failingExternalCmd+")[reason]").Puts(false),

		That("exc:is-pipeline-err ?(fail bad)[reason]").Puts(false),
		That("exc:is-pipeline-err ?(fail 1 | fail 2)[reason]").Puts(true),
	)
}
