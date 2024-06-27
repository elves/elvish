package prog_test

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/prog/progtest"
)

//go:embed *.elvts
var transcripts embed.FS

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"show-deprecation-level-in-global", func(ev *eval.Evaler) {
			ev.ExtendGlobal(eval.BuildNs().AddGoFn(
				"show-deprecation-level",
				func() int { return prog.DeprecationLevel }))
		},
		"program-makers-in-global", func(ev *eval.Evaler) {
			ev.ExtendGlobal(eval.BuildNs().
				AddGoFn("program-to-fn", programToFn).
				AddGoFn("composite", prog.Composite).
				AddGoFn("make-program", makeProgram).
				// Error constructors
				AddGoFn("next-program", nextProgram).
				AddGoFn("exit-error", func(i int) any { return prog.Exit(i) }).
				AddGoFn("bad-usage", func(s string) any { return prog.BadUsage(s) }))
		})
}

func programToFn(p prog.Program) eval.Callable {
	return eval.NewGoFn("test-program", progtest.ProgramAsGoFn(p))
}

type makeProgramOpts struct {
	WriteStdout string
	CustomFlag  bool
	SharedFlags bool
	ReturnErr   any
}

func (makeProgramOpts) SetDefaultOptions() {}

func makeProgram(opts makeProgramOpts) prog.Program {
	var returnErr error
	switch e := opts.ReturnErr.(type) {
	case error:
		returnErr = e
	case string:
		returnErr = errors.New(e)
	case nil:
		// Do nothing
	default:
		panic("&return-err should be error or string")
	}
	return &testProgram{
		writeStdout: opts.WriteStdout,
		customFlag:  opts.CustomFlag,
		sharedFlags: opts.SharedFlags,
		returnErr:   returnErr,
	}
}

func nextProgram(cleanupPrint ...string) any {
	switch len(cleanupPrint) {
	case 0:
		return prog.NextProgram()
	case 1:
		return prog.NextProgram(func(fds [3]*os.File) {
			fds[1].WriteString(cleanupPrint[0])
		})
	default:
		panic("next-program takes 0 or 1 argument")
	}
}

type testProgram struct {
	writeStdout string
	customFlag  bool
	sharedFlags bool
	returnErr   error

	flag        string
	daemonPaths *prog.DaemonPaths
	json        *bool
}

func (p *testProgram) RegisterFlags(f *prog.FlagSet) {
	if p.customFlag {
		f.StringVar(&p.flag, "flag", "default", "a flag")
	}
	if p.sharedFlags {
		p.daemonPaths = f.DaemonPaths()
		p.json = f.JSON()
	}
}

func (p *testProgram) Run(fds [3]*os.File, args []string) error {
	if p.returnErr != nil {
		return p.returnErr
	}
	fds[1].WriteString(p.writeStdout)
	if p.customFlag {
		fmt.Fprintf(fds[1], "-flag %s\n", p.flag)
	}
	if p.sharedFlags {
		fmt.Fprintf(fds[1], "-sock %s -db %s -json %v\n",
			p.daemonPaths.Sock, p.daemonPaths.DB, *p.json)
	}
	return nil
}
