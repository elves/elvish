package logutil

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"src.elv.sh/pkg/must"
)

func TestLogger(t *testing.T) {
	logger := GetLogger("foo ")

	r, w := must.Pipe()
	SetOutput(w)
	logger.Println("out 1")
	w.Close()
	wantOut1 := must.OK1(regexp.Compile("^foo .*out 1\n$"))
	if out := must.ReadAllAndClose(r); !wantOut1.Match(out) {
		t.Errorf("got out %q, want one matching %q", out, wantOut1)
	}

	outPath := filepath.Join(t.TempDir(), "out")
	must.OK(SetOutputFile(outPath))
	logger.Println("out 2")
	must.OK(SetOutputFile(""))
	wantOut2 := must.OK1(regexp.Compile("^foo .*out 2\n$"))
	if out := must.ReadAllAndClose(must.OK1(os.Open(outPath))); !wantOut2.Match(out) {
		t.Errorf("got out %q, want one matching %q", out, wantOut2)
	}
}

func TestSetOutput_Error(t *testing.T) {
	err := SetOutputFile("/bad/file/path")
	if err == nil {
		t.Errorf("want non-nil error, got nil")
	}
}
