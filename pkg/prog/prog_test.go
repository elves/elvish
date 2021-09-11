package prog_test

import (
	"os"
	"testing"

	. "src.elv.sh/pkg/prog"
	. "src.elv.sh/pkg/prog/progtest"
)

func TestBadFlag(t *testing.T) {
	f := Setup(t)

	exit := Run(f.Fds(), Elvish("-bad-flag"), testProgram{})

	TestError(t, f, exit, "flag provided but not defined: -bad-flag\nUsage:")
}

func TestDashHTreatedAsBadFlag(t *testing.T) {
	f := Setup(t)

	exit := Run(f.Fds(), Elvish("-h"), testProgram{})

	TestError(t, f, exit, "flag provided but not defined: -h\nUsage:")
}

func TestCPUProfile(t *testing.T) {
	f := Setup(t)

	Run(f.Fds(), Elvish("-cpuprofile", "cpuprof"), testProgram{})
	// There isn't much to test beyond a sanity check that the profile file now
	// exists.
	_, err := os.Stat("cpuprof")
	if err != nil {
		t.Errorf("CPU profile file does not exist: %v", err)
	}
}

func TestCPUProfile_BadPath(t *testing.T) {
	f := Setup(t)

	Run(f.Fds(), Elvish("-cpuprofile", "/a/bad/path"), testProgram{})
	f.TestOut(t, 1, "")
	f.TestOutSnippet(t, 2, "Warning: cannot create CPU profile:")
	f.TestOutSnippet(t, 2, "Continuing without CPU profiling.")
}

func TestHelp(t *testing.T) {
	f := Setup(t)

	Run(f.Fds(), Elvish("-help"), testProgram{})

	f.TestOutSnippet(t, 1, "Usage: elvish [flags] [script]")
	f.TestOut(t, 2, "")
}

func TestShowDeprecations(t *testing.T) {
	SetDeprecationLevel(t, 0)
	f := Setup(t)

	Run(f.Fds(), Elvish("-deprecation-level", "42"), testProgram{})
	if DeprecationLevel != 42 {
		t.Errorf("ShowDeprecations = %d, want 42", DeprecationLevel)
	}
}

func TestNoSuitableSubprogram(t *testing.T) {
	f := Setup(t)

	exit := Run(f.Fds(), Elvish(), testProgram{notSuitable: true})

	TestError(t, f, exit, "internal error: no suitable subprogram")
}

func TestComposite(t *testing.T) {
	f := Setup(t)

	Run(f.Fds(), Elvish(),
		Composite(testProgram{notSuitable: true}, testProgram{writeOut: "program 2"}))

	f.TestOut(t, 1, "program 2")
	f.TestOut(t, 2, "")
}

func TestComposite_PreferEarlierSubprogram(t *testing.T) {
	f := Setup(t)

	Run(f.Fds(), Elvish(),
		Composite(
			testProgram{writeOut: "program 1"}, testProgram{writeOut: "program 2"}))

	f.TestOut(t, 1, "program 1")
	f.TestOut(t, 2, "")
}

func TestBadUsageError(t *testing.T) {
	f := Setup(t)

	exit := Run(f.Fds(), Elvish(), testProgram{returnErr: BadUsage("lorem ipsum")})

	TestError(t, f, exit, "lorem ipsum")
	f.TestOutSnippet(t, 2, "Usage:")
	f.TestOut(t, 1, "")
}

func TestExitError(t *testing.T) {
	f := Setup(t)

	exit := Run(f.Fds(), Elvish(), testProgram{returnErr: Exit(3)})

	if exit != 3 {
		t.Errorf("exit = %v, want 3", exit)
	}
	f.TestOut(t, 2, "")
	f.TestOut(t, 1, "")
}

func TestExitError_0(t *testing.T) {
	f := Setup(t)

	exit := Run(f.Fds(), Elvish(), testProgram{returnErr: Exit(0)})

	if exit != 0 {
		t.Errorf("exit = %v, want 0", exit)
	}
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
