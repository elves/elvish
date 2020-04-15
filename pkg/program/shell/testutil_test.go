package shell

import (
	"io/ioutil"
	"os"

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
