package shell

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

type fixture struct {
	pipes      [3]*pipe
	dirCleanup func()
}

func setup() *fixture {
	_, dirCleanup := util.InTestDir()
	return &fixture{[3]*pipe{makePipe(), makePipe(), makePipe()}, dirCleanup}
}

func (f *fixture) cleanup() {
	f.pipes[0].close()
	f.pipes[1].close()
	f.pipes[2].close()
	f.dirCleanup()
}

func (f *fixture) fds() [3]*os.File {
	return [3]*os.File{f.pipes[0].r, f.pipes[1].w, f.pipes[2].w}
}

func (f *fixture) feedIn(s string) {
	_, err := f.pipes[0].w.WriteString(s)
	if err != nil {
		panic(err)
	}
	f.pipes[0].w.Close()
}

func (f *fixture) testOut(t *testing.T, fd int, wantOut string) {
	t.Helper()
	if out := f.pipes[fd].get(); out != wantOut {
		t.Errorf("got out %q, want %q", out, wantOut)
	}
}

func (f *fixture) testOutSnippet(t *testing.T, fd int, wantOutSnippet string) {
	t.Helper()
	if err := f.pipes[fd].get(); !strings.Contains(err, wantOutSnippet) {
		t.Errorf("got out %q, want string containing %q", err, wantOutSnippet)
	}
}

type pipe struct {
	r, w             *os.File
	rClosed, wClosed bool
	saved            string
}

func makePipe() *pipe {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return &pipe{r: r, w: w}
}

func (p *pipe) get() string {
	if p.rClosed {
		return p.saved
	}
	p.w.Close()
	p.wClosed = true
	b, err := ioutil.ReadAll(p.r)
	if err != nil {
		panic(err)
	}
	p.r.Close()
	p.rClosed = true
	p.saved = string(b)
	return p.saved
}

func (p *pipe) close() {
	if !p.wClosed {
		p.w.Close()
		p.wClosed = true
	}
	if !p.rClosed {
		p.r.Close()
		p.rClosed = true
	}
}

func writeFile(name, content string) {
	err := ioutil.WriteFile(name, []byte(content), 0600)
	if err != nil {
		panic(err)
	}
}
