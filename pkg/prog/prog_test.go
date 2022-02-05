package prog_test

import (
	"fmt"
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

func TestFlagHandling(t *testing.T) {
	Test(t, &testProgram{},
		ThatElvish("-bad-flag").
			ExitsWith(2).
			WritesStderrContaining("flag provided but not defined: -bad-flag\nUsage:"),
		// -h is treated as a bad flag
		ThatElvish("-h").
			ExitsWith(2).
			WritesStderrContaining("flag provided but not defined: -h\nUsage:"),

		ThatElvish("-help").
			WritesStdoutContaining("Usage: elvish [flags] [script]"),
	)
}

func TestLogFlag(t *testing.T) {
	testutil.InTempDir(t)
	Test(t, &testProgram{},
		ThatElvish("-log", "log").DoesNothing())
	_, err := os.Stat("log")
	if err != nil {
		t.Errorf("log file was not created: %v", err)
	}
}

func TestCustomFlag(t *testing.T) {
	Test(t, &testProgram{customFlag: true},
		ThatElvish("-flag", "foo").
			WritesStdout("-flag foo\n"),
	)
}

func TestSharedFlags(t *testing.T) {
	Test(t, &testProgram{sharedFlags: true},
		ThatElvish("-sock", "sock", "-db", "db", "-json").
			WritesStdout("-sock sock -db db -json true\n"),
	)
}

func TestSharedFlags_MultiplePrograms(t *testing.T) {
	Test(t,
		Composite(
			&testProgram{sharedFlags: true, nextProgram: true},
			&testProgram{sharedFlags: true}),
		ThatElvish("-sock", "sock", "-db", "db", "-json").
			WritesStdout("-sock sock -db db -json true\n"),
	)
}

func TestShowDeprecations(t *testing.T) {
	progtest.SetDeprecationLevel(t, 0)

	Test(t, &testProgram{},
		ThatElvish("-deprecation-level", "42").DoesNothing(),
	)

	if DeprecationLevel != 42 {
		t.Errorf("ShowDeprecations = %d, want 42", DeprecationLevel)
	}
}

func TestNoSuitableSubprogram(t *testing.T) {
	Test(t, &testProgram{nextProgram: true},
		ThatElvish().
			ExitsWith(2).
			WritesStderr("internal error: no suitable subprogram\n"),
	)
}

func TestComposite(t *testing.T) {
	Test(t,
		Composite(&testProgram{nextProgram: true}, &testProgram{writeOut: "program 2"}),
		ThatElvish().WritesStdout("program 2"),
	)
}

func TestComposite_NoSuitableSubprogram(t *testing.T) {
	Test(t,
		Composite(&testProgram{nextProgram: true}, &testProgram{nextProgram: true}),
		ThatElvish().
			ExitsWith(2).
			WritesStderr("internal error: no suitable subprogram\n"),
	)
}

func TestComposite_PreferEarlierSubprogram(t *testing.T) {
	Test(t,
		Composite(
			&testProgram{writeOut: "program 1"}, &testProgram{writeOut: "program 2"}),
		ThatElvish().WritesStdout("program 1"),
	)
}

func TestBadUsageError(t *testing.T) {
	Test(t,
		&testProgram{returnErr: BadUsage("lorem ipsum")},
		ThatElvish().ExitsWith(2).WritesStderrContaining("lorem ipsum\n"),
	)
}

func TestExitError(t *testing.T) {
	Test(t, &testProgram{returnErr: Exit(3)},
		ThatElvish().ExitsWith(3),
	)
}

func TestExitError_0(t *testing.T) {
	Test(t, &testProgram{returnErr: Exit(0)},
		ThatElvish().ExitsWith(0),
	)
}

type testProgram struct {
	nextProgram bool
	writeOut    string
	returnErr   error
	customFlag  bool
	sharedFlags bool

	flag        string
	daemonPaths *DaemonPaths
	json        *bool
}

func (p *testProgram) RegisterFlags(f *FlagSet) {
	if p.customFlag {
		f.StringVar(&p.flag, "flag", "default", "a flag")
	}
	if p.sharedFlags {
		p.daemonPaths = f.DaemonPaths()
		p.json = f.JSON()
	}
}

func (p *testProgram) Run(fds [3]*os.File, args []string) error {
	if p.nextProgram {
		return ErrNextProgram
	}
	fds[1].WriteString(p.writeOut)
	if p.customFlag {
		fmt.Fprintf(fds[1], "-flag %s\n", p.flag)
	}
	if p.sharedFlags {
		fmt.Fprintf(fds[1], "-sock %s -db %s -json %v\n",
			p.daemonPaths.Sock, p.daemonPaths.DB, *p.json)
	}
	return p.returnErr
}
