package file

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/evaltest"
)

func TestFile(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("file", Ns).Ns())
	}
	evaltest.TestWithSetup(t, setup,
		evaltest.That(`echo haha > out3`, `f = (file:open out3)`, `slurp <$f`, ` file:close $f`).Puts("haha\n"),
		//No name is specified hence should give an error.
		evaltest.That(`file:open`).Throws(evaltest.AnyError),
		evaltest.That(`file:close`).Throws(evaltest.AnyError),

		evaltest.That(`file:close`).Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 1, ValidHigh: 1, Actual: 0}),

		evaltest.That(`file:open`).Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 1, ValidHigh: 1, Actual: 0}),

		//Should given an error when try to close an already closed file.
		evaltest.That(`echo haha > out3`, `f = (file:open out3)`, ` file:close $f`, `file:close $f`).Throws(evaltest.AnyError),
	)
}
