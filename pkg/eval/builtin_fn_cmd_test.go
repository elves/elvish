package eval_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestExit(t *testing.T) {
	var exitCodes []int
	testutil.Set(t, OSExit, func(i int) { exitCodes = append(exitCodes, i) })

	Test(t,
		That("exit").DoesNothing(),
		That("exit 1").DoesNothing(),
		That("exit 1 2").Throws(
			errs.ArityMismatch{What: "arguments",
				ValidLow: 0, ValidHigh: 1, Actual: 2},
			"exit 1 2"),
	)

	if diff := cmp.Diff([]int{0, 1}, exitCodes); diff != "" {
		t.Errorf("got unexpected exit codes (-want +got):\n%s", diff)
	}
}

func TestExit_RunsPreExit(t *testing.T) {
	testutil.Set(t, OSExit, func(int) {})

	calls := 0
	setup := func(ev *Evaler) {
		ev.PreExitHooks = append(ev.PreExitHooks, func() { calls++ })
	}

	TestWithEvalerSetup(t, setup,
		That("exit").DoesNothing())

	if calls != 1 {
		t.Errorf("pre-exit hook called %v times, want 1", calls)
	}
}
