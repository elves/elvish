package eval

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestEvalerPorts(t *testing.T) {
	stdoutReader, stdout := mustPipe()
	defer stdoutReader.Close()

	stderrReader, stderr := mustPipe()
	defer stderrReader.Close()

	prefix := "> "
	ep := newEvalerPorts(DevNull, stdout, stderr, &prefix)
	ep.ports[1].Chan <- "x"
	ep.ports[1].Chan <- "y"
	ep.ports[2].Chan <- "bad"
	ep.ports[2].Chan <- "err"
	ep.close()
	stdout.Close()
	stderr.Close()

	stdoutAll := mustReadAllString(stdoutReader)
	wantStdoutAll := "> x\n> y\n"
	if stdoutAll != wantStdoutAll {
		t.Errorf("stdout is %q, want %q", stdoutAll, wantStdoutAll)
	}
	stderrAll := mustReadAllString(stderrReader)
	wantStderrAll := "> bad\n> err\n"
	if stderrAll != wantStderrAll {
		t.Errorf("stderr is %q, want %q", stderrAll, wantStderrAll)
	}
}

func mustPipe() (*os.File, *os.File) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return r, w
}

func mustReadAllString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}
