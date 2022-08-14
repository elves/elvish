package eval_test

import (
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/prog"
)

func TestExit(t *testing.T) {
	Test(t,
		That("exit").PanicsWith(prog.ExitStatus{Status: 0}),
		That("exit 1").PanicsWith(prog.ExitStatus{Status: 1}),
		That("exit 256").Throws(
			errs.OutOfRange{
				What:     "exit code",
				ValidLow: "0", ValidHigh: "255", Actual: "256",
			}),
		That("exit -1").Throws(
			errs.OutOfRange{
				What:     "exit code",
				ValidLow: "0", ValidHigh: "255", Actual: "-1",
			}),
		That("exit 1 2").Throws(
			errs.ArityMismatch{What: "arguments",
				ValidLow: 0, ValidHigh: 1, Actual: 2},
			"exit 1 2"),
	)
}

func TestExit_RunsPreExit(t *testing.T) {
	calls := 0
	setup := func(ev *Evaler) {
		ev.PreExitHooks = append(ev.PreExitHooks, func() { calls++ })
	}

	TestWithSetup(t, setup,
		That("exit").PanicsWith(prog.ExitStatus{Status: 0}),
	)

	if calls != 1 {
		t.Errorf("pre-exit hook called %v times, want 1", calls)
	}
}
