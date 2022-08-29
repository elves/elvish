// Package progtest provides a framework for testing subprograms.
//
// The entry point for the framework is the Test function, which accepts a
// *testing.T, the Program implementation under test, and any number of test
// cases.
//
// Test cases are constructed using the ThatElvish function, followed by method
// calls that add additional information to it.
//
// Example:
//
//	Test(t, someProgram,
//	     ThatElvish("-c", "echo hello").WritesStdout("hello\n"))
package progtest

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/prog"
)

// Case is a test case that can be used in Test.
type Case struct {
	args  []string
	stdin string
	want  result
}

type result struct {
	exitCode int
	stdout   output
	stderr   output
}

type output struct {
	content string
	partial bool
}

func (o output) String() string {
	if o.partial {
		return fmt.Sprintf("text containing %q", o.content)
	}
	return fmt.Sprintf("%q", o.content)
}

// ThatElvish returns a new Case with the specified CLI arguments.
//
// The new Case expects the program run to exit with 0, and write nothing to
// stdout or stderr.
//
// When combined with subsequent method calls, a test case reads like English.
// For example, a test for the fact that "elvish -c hello" writes "hello\n" to
// stdout reads:
//
//	ThatElvish("-c", "hello").WritesStdout("hello\n")
func ThatElvish(args ...string) Case {
	return Case{args: append([]string{"elvish"}, args...)}
}

// WithStdin returns an altered Case that provides the given input to stdin of
// the program.
func (c Case) WithStdin(s string) Case {
	c.stdin = s
	return c
}

// DoesNothing returns c itself. It is useful to mark tests that otherwise don't
// have any expectations, for example:
//
//	ThatElvish("-c", "nop").DoesNothing()
func (c Case) DoesNothing() Case {
	return c
}

// ExitsWith returns an altered Case that requires the program run to return
// with the given exit code.
func (c Case) ExitsWith(code int) Case {
	c.want.exitCode = code
	return c
}

// WritesStdout returns an altered Case that requires the program run to write
// exactly the given text to stdout.
func (c Case) WritesStdout(s string) Case {
	c.want.stdout = output{content: s}
	return c
}

// WritesStdoutContaining returns an altered Case that requires the program run
// to write output to stdout that contains the given text as a substring.
func (c Case) WritesStdoutContaining(s string) Case {
	c.want.stdout = output{content: s, partial: true}
	return c
}

// WritesStderr returns an altered Case that requires the program run to write
// exactly the given text to stderr.
func (c Case) WritesStderr(s string) Case {
	c.want.stderr = output{content: s}
	return c
}

// WritesStderrContaining returns an altered Case that requires the program run
// to write output to stderr that contains the given text as a substring.
func (c Case) WritesStderrContaining(s string) Case {
	c.want.stderr = output{content: s, partial: true}
	return c
}

// Test runs test cases against a given program.
func Test(t *testing.T, p prog.Program, cases ...Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			t.Helper()
			r := run(p, c.args, c.stdin)
			if r.exitCode != c.want.exitCode {
				t.Errorf("got exit code %v, want %v", r.exitCode, c.want.exitCode)
			}
			if !matchOutput(r.stdout, c.want.stdout) {
				t.Errorf("got stdout %v, want %v", r.stdout, c.want.stdout)
			}
			if !matchOutput(r.stderr, c.want.stderr) {
				t.Errorf("got stderr %v, want %v", r.stderr, c.want.stderr)
			}
		})
	}
}

// Run runs a Program with the given arguments. It returns the Program's exit
// code and output to stdout and stderr.
func Run(p prog.Program, args ...string) (exit int, stdout, stderr string) {
	r := run(p, args, "")
	return r.exitCode, r.stdout.content, r.stderr.content
}

func run(p prog.Program, args []string, stdin string) result {
	r0, w0 := must.Pipe()
	// TODO: This assumes that stdin fits in the pipe buffer. Don't assume that.
	_, err := w0.WriteString(stdin)
	if err != nil {
		panic(err)
	}
	w0.Close()
	defer r0.Close()

	w1, get1 := capturedOutput()
	w2, get2 := capturedOutput()

	exitCode := prog.Run([3]*os.File{r0, w1, w2}, args, p)
	return result{exitCode, output{content: get1()}, output{content: get2()}}
}

func matchOutput(got, want output) bool {
	if want.partial {
		return strings.Contains(got.content, want.content)
	}
	return got.content == want.content
}

func capturedOutput() (*os.File, func() string) {
	r, w := must.Pipe()
	output := make(chan string, 1)
	go func() {
		b, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}
		r.Close()
		output <- string(b)
	}()
	return w, func() string {
		// Close the write side so captureOutput goroutine sees EOF and
		// terminates allowing us to capture and cache the output.
		w.Close()
		return <-output
	}
}
