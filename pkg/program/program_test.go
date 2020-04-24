package program

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/buildinfo"
)

func parse(args []string) (*flag.FlagSet, *Flags) {
	f := Flags{}
	fs := newFlagSet(os.Stderr, &f)
	err := fs.Parse(args)
	if err != nil {
		panic(fmt.Sprintln("bad flags in test", args))
	}
	return fs, &f
}

type programTest struct {
	run   []string
	stdin string

	wantExit         int
	wantStdout       string
	wantStderr       string
	wantStdoutPrefix string
	wantStderrPrefix string
}

var programTests = []programTest{
	{
		run:        elvish("-version"),
		wantStdout: buildinfo.Version + "\n",
	},
	{
		run: elvish("-buildinfo"),
		wantStdout: fmt.Sprintf(
			"Version: %v\nGo version: %v\nReproducible build: %v\n",
			buildinfo.Version,
			runtime.Version(),
			buildinfo.Reproducible,
		),
	},
	{
		run: elvish("-buildinfo", "-json"),
		wantStdout: mustToJSON(struct {
			Version      string `json:"version"`
			GoVersion    string `json:"goversion"`
			Reproducible bool   `json:"reproducible"`
		}{
			buildinfo.Version,
			runtime.Version(),
			buildinfo.Reproducible == "true",
		}) + "\n",
	},
	{
		run:              elvish("-help"),
		wantStdoutPrefix: "Usage: elvish [flags] [script]",
	},
	// Bad usages.
	badUsage("flag provided but not defined: -bad-flag", "-bad-flag"),
	badUsage("arguments are not allowed with -web", "-web", "x"),
	badUsage("-c cannot be used together with -web", "-web", "-c"),
	badUsage("arguments are not allowed with -daemon", "-daemon", "x"),
}

func elvish(args ...string) []string {
	return append([]string{"elvish"}, args...)
}

func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func badUsage(prefix string, args ...string) programTest {
	return programTest{
		run:              elvish(args...),
		wantStderrPrefix: prefix + "\nUsage: elvish [flags] [script]",
		wantExit:         2,
	}
}

func TestPrograms(t *testing.T) {
	for _, test := range programTests {
		t.Run(strings.Join(test.run, " "), func(t *testing.T) {
			fd0, fd0w := mustMakePipe()
			fd1r, fd1 := mustMakePipe()
			fd2r, fd2 := mustMakePipe()

			fd0w.WriteString(test.stdin)
			fd0w.Close()

			exit := Main([3]*os.File{fd0, fd1, fd2}, test.run)
			fd0.Close()
			fd1.Close()
			fd2.Close()

			stdout := mustReadAllAndClose(fd1r)
			stderr := mustReadAllAndClose(fd2r)
			testOutput(t, "stdout", stdout, test.wantStdout, test.wantStdoutPrefix)
			testOutput(t, "stderr", stderr, test.wantStderr, test.wantStderrPrefix)
			if exit != test.wantExit {
				t.Errorf("got exit %d, want %d", exit, test.wantExit)
			}
		})
	}
}

func testOutput(t *testing.T, name, got, want, wantPrefix string) {
	if wantPrefix != "" {
		if !strings.HasPrefix(got, wantPrefix) {
			t.Errorf("%s mismatch", name)
			t.Logf("got:\n%s", got)
			t.Logf("want prefix:\n%s", wantPrefix)
		}
	} else if got != want {
		t.Errorf("%s mismatch", name)
		t.Logf("got:\n%s", got)
		t.Logf("want:\n%s", want)
	}
}

func mustMakePipe() (r, w *os.File) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return r, w
}

func mustReadAllAndClose(r io.ReadCloser) string {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	r.Close()
	return string(b)
}
