package file

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
)

func TestFile(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("file", Ns).Ns())
	}
	evaltest.TestWithSetup(t, setup,
		evaltest.That(`file:fopen`).Throws(evaltest.AnyError),
	)
}
