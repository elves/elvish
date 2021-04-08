package file

import (
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestFile(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("file", Ns).Ns())
	}
	_, cleanup := testutil.InTestDir()
	defer cleanup()
	TestWithSetup(t, setup,
		That(
			"echo haha > out3", "f = (file:open out3)",
			"slurp < $f", "file:close $f").Puts("haha\n"),

		That("file:open").Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 1, ValidHigh: 1, Actual: 0}),

		That("file:close").Throws(
			errs.ArityMismatch{
				What:     "arguments here",
				ValidLow: 1, ValidHigh: 1, Actual: 0}),
	)

}
