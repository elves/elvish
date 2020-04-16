package shell

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/util"
)

type fixture struct {
	stdout     *pipedOut
	stderr     *pipedOut
	dirCleanup func()
}

func setup() *fixture {
	_, dirCleanup := util.InTestDir()
	return &fixture{makePipedOut(), makePipedOut(), dirCleanup}
}

func (f *fixture) fds() [3]*os.File {
	return [3]*os.File{eval.DevNull, f.stdout.w, f.stderr.w}
}

func (f *fixture) testOut(t *testing.T, wantOut string) {
	t.Helper()
	if out := f.getOut(); out != wantOut {
		t.Errorf("got out %q, want %q", out, wantOut)
	}
}

func (f *fixture) testOutSnippet(t *testing.T, wantOutSnippet string) {
	t.Helper()
	if err := f.getOut(); !strings.Contains(err, wantOutSnippet) {
		t.Errorf("got err %q, want string containing %q", err, wantOutSnippet)
	}
}

func (f *fixture) testErr(t *testing.T, wantErr string) {
	t.Helper()
	if err := f.getErr(); err != wantErr {
		t.Errorf("got err %q, want %q", err, wantErr)
	}
}

func (f *fixture) testErrSnippet(t *testing.T, wantErrSnippet string) {
	t.Helper()
	if err := f.getErr(); !strings.Contains(err, wantErrSnippet) {
		t.Errorf("got err %q, want string containing %q", err, wantErrSnippet)
	}
}

func (f *fixture) getOut() string { return f.stdout.get() }

func (f *fixture) getErr() string { return f.stderr.get() }

func (f *fixture) cleanup() {
	f.stdout.close()
	f.stderr.close()
	f.dirCleanup()
}

type pipedOut struct {
	r, w   *os.File
	closed bool
	saved  string
}

func makePipedOut() *pipedOut {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return &pipedOut{r: r, w: w}
}

func (p *pipedOut) get() string {
	if p.closed {
		return p.saved
	}
	p.w.Close()
	b, err := ioutil.ReadAll(p.r)
	if err != nil {
		panic(err)
	}
	p.r.Close()
	p.closed = true
	p.saved = string(b)
	return p.saved
}

func (p *pipedOut) close() {
	if !p.closed {
		p.w.Close()
		p.r.Close()
		p.closed = true
	}
}

func writeFile(name, content string) {
	err := ioutil.WriteFile(name, []byte(content), 0600)
	if err != nil {
		panic(err)
	}
}
