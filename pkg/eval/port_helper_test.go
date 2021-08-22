package eval

import (
	"io"
	"os"
	"testing"
)

func TestEvalerPorts(t *testing.T) {
	stdoutReader, stdout := mustPipe()
	defer stdoutReader.Close()

	stderrReader, stderr := mustPipe()
	defer stderrReader.Close()

	prefix := "> "
	ports, cleanup := PortsFromFiles([3]*os.File{DevNull, stdout, stderr}, prefix)
	ports[1].Chan <- "x"
	ports[1].Chan <- "y"
	ports[2].Chan <- "bad"
	ports[2].Chan <- "err"
	cleanup()
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

func mustReadAllString(r io.Reader) string {
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func mustPipe() (*os.File, *os.File) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return r, w
}
