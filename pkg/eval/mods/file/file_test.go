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
		evaltest.That(`f = (file:fopen file.go)`).Puts(),

		evaltest.That(`file:fclose`).Throws(evaltest.AnyError),
		evaltest.That(`file:fclose $f`).Puts(),

		evaltest.That(`p = (file:pipe)`).Puts(),
		evaltest.That(`file:prclose $p`).Puts(),
	)
}
