// Package progtest provides utilities for testing subprograms.
//
// This package intentionally has no test file; it is excluded from test
// coverage.
package progtest

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"src.elv.sh/pkg/testutil"
)

// Fixture is a test fixture suitable for testing programs.
type Fixture struct {
	pipes      [3]*pipe
	dirCleanup func()
}

func captureOutput(p *pipe) {
	b, err := ioutil.ReadAll(p.r)
	if err != nil {
		panic(err)
	}
	p.output <- b
}

// Setup sets up a test fixture. The caller is responsible for calling the
// Cleanup method of the returned Fixture.
func Setup() *Fixture {
	_, dirCleanup := testutil.InTestDir()
	pipes := [3]*pipe{makePipe(false), makePipe(true), makePipe(true)}
	return &Fixture{pipes, dirCleanup}
}

// Cleanup cleans up the test fixture.
func (f *Fixture) Cleanup() {
	f.pipes[0].close()
	f.pipes[1].close()
	f.pipes[2].close()
	f.dirCleanup()
}

// Fds returns the file descriptors in the fixture.
func (f *Fixture) Fds() [3]*os.File {
	return [3]*os.File{f.pipes[0].r, f.pipes[1].w, f.pipes[2].w}
}

// FeedIn feeds input to the standard input.
func (f *Fixture) FeedIn(s string) {
	_, err := f.pipes[0].w.WriteString(s)
	if err != nil {
		panic(err)
	}
	f.pipes[0].w.Close()
	f.pipes[0].wClosed = true
}

// TestOut tests that the output on the given FD matches the given text.
func (f *Fixture) TestOut(t *testing.T, fd int, wantOut string) {
	t.Helper()
	if out := f.pipes[fd].get(); out != wantOut {
		t.Errorf("got out %q, want %q", out, wantOut)
	}
}

// TestOutSnippet tests that the output on the given FD contains the given text.
func (f *Fixture) TestOutSnippet(t *testing.T, fd int, wantOutSnippet string) {
	t.Helper()
	if err := f.pipes[fd].get(); !strings.Contains(err, wantOutSnippet) {
		t.Errorf("got out %q, want string containing %q", err, wantOutSnippet)
	}
}

type pipe struct {
	r, w             *os.File
	rClosed, wClosed bool
	saved            string
	output           chan []byte
}

func makePipe(capture bool) *pipe {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	if !capture {
		return &pipe{r: r, w: w}
	}
	output := make(chan []byte, 1)
	p := pipe{r: r, w: w, output: output}
	go captureOutput(&p)
	return &p
}

func (p *pipe) get() string {
	if !p.wClosed {
		// Close the write side so captureOutput goroutine sees EOF and
		// terminates allowing us to capture and cache the output.
		p.w.Close()
		p.wClosed = true
		if p.output != nil {
			p.saved = string(<-p.output)
		}
	}
	return p.saved
}

func (p *pipe) close() {
	if !p.wClosed {
		p.w.Close()
		p.wClosed = true
		if p.output != nil {
			p.saved = string(<-p.output)
		}
	}
	if !p.rClosed {
		p.r.Close()
		p.rClosed = true
	}
	if p.output != nil {
		close(p.output)
		p.output = nil
	}
}

// MustWriteFile writes a file with the given name and content. It panics if the
// write fails.
func MustWriteFile(name, content string) {
	err := ioutil.WriteFile(name, []byte(content), 0600)
	if err != nil {
		panic(err)
	}
}

// Elvish returns an argument slice starting with "elvish".
func Elvish(args ...string) []string {
	return append([]string{"elvish"}, args...)
}

// TestError tests the error result of a program.
func TestError(t *testing.T, f *Fixture, exit int, wantErrSnippet string) {
	t.Helper()
	if exit != 2 {
		t.Errorf("got exit %v, want 2", exit)
	}
	f.TestOutSnippet(t, 2, wantErrSnippet)
}
