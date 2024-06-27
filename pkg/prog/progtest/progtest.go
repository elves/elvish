// Package progtest contains utilities for wrapping [prog.Program] instances
// into Elvish functions, so that they can be tested using the
// [src.elv.sh/pkg/eval/evaltest] framework.
package progtest

import (
	"fmt"
	"io"
	"os"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/prog"
)

// ElvishInGlobal returns a setup function suitable for the evaltest framework,
// which creates a function called "elvish" in the global scope that invokes the
// given program.
func ElvishInGlobal(p prog.Program) func(ev *eval.Evaler) {
	return func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddGoFn("elvish", ProgramAsGoFn(p)))
	}
}

type programOpts struct {
	CheckStdoutContains string
	CheckStderrContains string
}

func (programOpts) SetDefaultOptions() {}

// ProgramAsGoFn converts a [prog.Program] to a Go-implemented Elvish function.
//
// Stdin of the program is connected to the stdin of the function.
//
// Stdout of the program is usually written unchanged to the stdout of the
// function, except when:
//
//   - If the output has no trailing newline, " (no EOL)\n" is appended.
//   - If &check-stdout-contains is supplied, stdout is suppressed. Instead, a
//     tag "[stdout contains foo]" is shown, followed by "true" or "false".
//
// Stderr of the program is written to the stderr of the function with a
// [stderr] prefix, with similar treatment for missing trailing EOL and
// &check-stderr-contains.
//
// If the program exits with a non-zero return value, a line "[exit] $i" is
// written to stderr.
func ProgramAsGoFn(p prog.Program) any {
	return func(fm *eval.Frame, opts programOpts, args ...string) {
		r1, w1 := must.OK2(os.Pipe())
		r2, w2 := must.OK2(os.Pipe())
		args = append([]string{"elvish"}, args...)
		exit := prog.Run([3]*os.File{fm.InputFile(), w1, w2}, args, p)
		w1.Close()
		w2.Close()

		outFile := fm.ByteOutput()
		stdout := string(must.OK1(io.ReadAll(r1)))
		if s := opts.CheckStdoutContains; s != "" {
			fmt.Fprintf(outFile,
				"[stdout contains %q] %t\n", s, strings.Contains(stdout, s))
		} else {
			outFile.WriteString(lines("", stdout))
		}

		errFile := fm.ErrorFile()
		stderr := string(must.OK1(io.ReadAll(r2)))
		if s := opts.CheckStderrContains; s != "" {
			fmt.Fprintf(errFile,
				"[stderr contains %q] %t\n", s, strings.Contains(stderr, s))
		} else {
			errFile.WriteString(lines("[stderr] ", stderr))
		}

		if exit != 0 {
			fmt.Fprintf(errFile, "[exit] %d\n", exit)
		}
	}
}

// Splits data into lines, adding prefix to each line and appending " (no EOL)"
// if data doesn't end in a newline.
func lines(prefix, data string) string {
	var sb strings.Builder
	for len(data) > 0 {
		sb.WriteString(prefix)
		i := strings.IndexByte(data, '\n')
		if i == -1 {
			sb.WriteString(data + " (no EOL)\n")
			break
		} else {
			sb.WriteString(data[:i+1])
			data = data[i+1:]
		}
	}
	return sb.String()
}
