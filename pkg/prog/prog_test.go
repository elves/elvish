package prog_test

import (
	"os"
	"testing"

	. "src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

var (
	Test       = progtest.Test
	ThatElvish = progtest.ThatElvish
)

func TestCommonFlagHandling(t *testing.T) {
	testutil.InTempDir(t)

	Test(t, testProgram{},
		ThatElvish("-bad-flag").
			ExitsWith(2).
			WritesStderrContaining("flag provided but not defined: -bad-flag\nUsage:"),
		// -h is treated as a bad flag
		ThatElvish("-h").
			ExitsWith(2).
			WritesStderrContaining("flag provided but not defined: -h\nUsage:"),

		ThatElvish("-help").
			WritesStdoutContaining("Usage: elvish [flags] [script]"),

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

func TestShowDeprecations(t *testing.T) {
	progtest.SetDeprecationLevel(t, 0)

	Test(t, testProgram{},
		ThatElvish("-deprecation-level", "42").DoesNothing(),
	)

	if DeprecationLevel != 42 {
		t.Errorf("ShowDeprecations = %d, want 42", DeprecationLevel)
	}
}

func TestNoSuitableSubprogram(t *testing.T) {
	Test(t, testProgram{notSuitable: true},
		ThatElvish().
			ExitsWith(2).
			WritesStderr("internal error: no suitable subprogram\n"),
	)
}

func TestComposite(t *testing.T) {
	Test(t,
		Composite(testProgram{notSuitable: true}, testProgram{writeOut: "program 2"}),
		ThatElvish().WritesStdout("program 2"),
	)
}

func TestComposite_NoSuitableSubprogram(t *testing.T) {
	Test(t,
		Composite(testProgram{notSuitable: true}, testProgram{notSuitable: true}),
		ThatElvish().
			ExitsWith(2).
			WritesStderr("internal error: no suitable subprogram\n"),
	)
}

func TestComposite_PreferEarlierSubprogram(t *testing.T) {
	Test(t,
		Composite(
			testProgram{writeOut: "program 1"}, testProgram{writeOut: "program 2"}),
		ThatElvish().WritesStdout("program 1"),
	)
}

func TestBadUsageError(t *testing.T) {
	Test(t,
		testProgram{returnErr: BadUsage("lorem ipsum")},
		ThatElvish().ExitsWith(2).WritesStderrContaining("lorem ipsum\n"),
	)
}

func TestExitError(t *testing.T) {
	Test(t, testProgram{returnErr: Exit(3)},
		ThatElvish().ExitsWith(3),
	)
}

func TestExitError_0(t *testing.T) {
	Test(t, testProgram{returnErr: Exit(0)},
		ThatElvish().ExitsWith(0),
	)
}

type testProgram struct {
	notSuitable bool
	writeOut    string
	returnErr   error
}

func (p testProgram) Run(fds [3]*os.File, _ *Flags, args []string) error {
	if p.notSuitable {
		return ErrNotSuitable
	}
	fds[1].WriteString(p.writeOut)
	return p.returnErr
}
