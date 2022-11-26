package pprof_test

import (
	"os"
	"testing"

	. "src.elv.sh/pkg/pprof"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

var (
	Test       = progtest.Test
	ThatElvish = progtest.ThatElvish
)

func TestProgram(t *testing.T) {
	testutil.InTempDir(t)

	Test(t, prog.Composite(&Program{}, noopProgram{}),
		ThatElvish("-cpuprofile", "cpu").DoesNothing(),
		ThatElvish("-cpuprofile", "bad/path").
			WritesStderrContaining("Warning: cannot create CPU profile:"),

		ThatElvish("-allocsprofile", "allocs").DoesNothing(),
		ThatElvish("-allocsprofile", "bad/path").
			WritesStderrContaining("Warning: cannot create memory allocation profile:"),
	)

	// Check for the effects of -cpuprofile and -allocsprofile. There isn't much
	// that can be checked easily, so we only do a sanity check that the profile
	// files exist and are non-empty.
	checkFileIsNonEmpty(t, "cpu")
	checkFileIsNonEmpty(t, "allocs")
}

func checkFileIsNonEmpty(t *testing.T, name string) {
	t.Helper()
	stat, err := os.Stat(name)
	if err != nil {
		t.Errorf("CPU profile file does not exist: %v", err)
	} else if stat.Size() == 0 {
		t.Errorf("CPU profile exists but is empty")
	}
}

type noopProgram struct{}

func (noopProgram) RegisterFlags(*prog.FlagSet)     {}
func (noopProgram) Run([3]*os.File, []string) error { return nil }
