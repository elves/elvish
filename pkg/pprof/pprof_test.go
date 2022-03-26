package pprof_test

import (
	"os"
	"testing"

	"src.elv.sh/pkg/pprof"
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

	Test(t, prog.Composite(&pprof.Program{}, noopProgram{}),
		ThatElvish("-cpuprofile", "cpuprof").DoesNothing(),
		ThatElvish("-cpuprofile", "/a/bad/path").
			WritesStderrContaining("Warning: cannot create CPU profile:"),
	)

	// Check for the effect of -cpuprofile. There isn't much to test beyond a
	// sanity check that the profile file now exists.
	_, err := os.Stat("cpuprof")
	if err != nil {
		t.Errorf("CPU profile file does not exist: %v", err)
	}
}

type noopProgram struct{}

func (noopProgram) RegisterFlags(*prog.FlagSet)     {}
func (noopProgram) Run([3]*os.File, []string) error { return nil }
