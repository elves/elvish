package prog

import (
	"os"
	"testing"

	. "src.elv.sh/pkg/prog/progtest"
)

func TestBadFlag(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish("-bad-flag"))

	TestError(t, f, exit, "flag provided but not defined: -bad-flag\nUsage:")
}

func TestDashHTreatedAsBadFlag(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish("-h"))

	TestError(t, f, exit, "flag provided but not defined: -h\nUsage:")
}

func TestCPUProfile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish("-cpuprofile", "cpuprof"), testProgram{shouldRun: true})
	// There isn't much to test beyond a sanity check that the profile file now
	// exists.
	_, err := os.Stat("cpuprof")
	if err != nil {
		t.Errorf("CPU profile file does not exist: %v", err)
	}
}

func TestCPUProfile_BadPath(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish("-cpuprofile", "/a/bad/path"), testProgram{shouldRun: true})
	f.TestOut(t, 1, "")
	f.TestOutSnippet(t, 2, "Warning: cannot create CPU profile:")
	f.TestOutSnippet(t, 2, "Continuing without CPU profiling.")
}

func TestHelp(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish("-help"))

	f.TestOutSnippet(t, 1, "Usage: elvish [flags] [script]")
	f.TestOut(t, 2, "")
}

func TestShowDeprecations(t *testing.T) {
	restore := SetDeprecationLevel(0)
	defer restore()
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish("-deprecation-level", "42"), testProgram{shouldRun: true})
	if DeprecationLevel != 42 {
		t.Errorf("ShowDeprecations = %d, want 42", DeprecationLevel)
	}
}

func TestNoProgram(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish(), testProgram{}, testProgram{})

	TestError(t, f, exit, "program bug: no suitable subprogram")
}

func TestGoodProgram(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish(), testProgram{},
		testProgram{shouldRun: true, writeOut: "program 2"})

	f.TestOut(t, 1, "program 2")
	f.TestOut(t, 2, "")
}

func TestPreferEarlierProgram(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Run(f.Fds(), Elvish(),
		testProgram{shouldRun: true, writeOut: "program 1"},
		testProgram{shouldRun: true, writeOut: "program 2"})

	f.TestOut(t, 1, "program 1")
	f.TestOut(t, 2, "")
}

func TestBadUsageError(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish(),
		testProgram{shouldRun: true, returnErr: BadUsage("lorem ipsum")})

	TestError(t, f, exit, "lorem ipsum")
	f.TestOutSnippet(t, 2, "Usage:")
	f.TestOut(t, 1, "")
}

func TestExitError(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish(),
		testProgram{shouldRun: true, returnErr: Exit(3)})

	if exit != 3 {
		t.Errorf("exit = %v, want 3", exit)
	}
	f.TestOut(t, 2, "")
	f.TestOut(t, 1, "")
}

func TestExitError_0(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	exit := Run(f.Fds(), Elvish(),
		testProgram{shouldRun: true, returnErr: Exit(0)})

	if exit != 0 {
		t.Errorf("exit = %v, want 0", exit)
	}
}

type testProgram struct {
	shouldRun bool
	writeOut  string
	returnErr error
}

func (p testProgram) ShouldRun(*Flags) bool { return p.shouldRun }

func (p testProgram) Run(fds [3]*os.File, _ *Flags, args []string) error {
	fds[1].WriteString(p.writeOut)
	return p.returnErr
}
