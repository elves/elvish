package program

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/buildinfo"
	"github.com/elves/elvish/pkg/program/shell"
	"github.com/elves/elvish/pkg/program/web"
)

var findProgramTests = []struct {
	args    []string
	checker func(Program) bool
}{
	{[]string{"-help"}, isShowHelp},
	{[]string{}, isShell},
	{[]string{"-c", "echo"}, func(p Program) bool {
		return p.(*shell.Shell).Cmd
	}},
	{[]string{"-compileonly"}, func(p Program) bool {
		return p.(*shell.Shell).CompileOnly
	}},
	{[]string{"-web"}, isWeb},
	{[]string{"-web", "x"}, isShowCorrectUsage},
	{[]string{"-web", "-c"}, isShowCorrectUsage},
	{[]string{"-web", "-port", "233"}, func(p Program) bool {
		return p.(*web.Web).Port == 233
	}},
	{[]string{"-daemon"}, isDaemon},
	{[]string{"-daemon", "x"}, isShowCorrectUsage},

	{[]string{"-bin", "/elvish"}, func(p Program) bool {
		return p.(*shell.Shell).BinPath == "/elvish"
	}},
	{[]string{"-db", "/db"}, func(p Program) bool {
		return p.(*shell.Shell).DbPath == "/db"
	}},
	{[]string{"-sock", "/sock"}, func(p Program) bool {
		return p.(*shell.Shell).SockPath == "/sock"
	}},

	{[]string{"-web", "-bin", "/elvish"}, func(p Program) bool {
		return p.(*web.Web).BinPath == "/elvish"
	}},
	{[]string{"-web", "-db", "/db"}, func(p Program) bool {
		return p.(*web.Web).DbPath == "/db"
	}},
	{[]string{"-web", "-sock", "/sock"}, func(p Program) bool {
		return p.(*web.Web).SockPath == "/sock"
	}},

	{[]string{"-daemon", "-bin", "/elvish"}, func(p Program) bool {
		return p.(Daemon).inner.BinPath == "/elvish"
	}},
	{[]string{"-daemon", "-db", "/db"}, func(p Program) bool {
		return p.(Daemon).inner.DbPath == "/db"
	}},
	{[]string{"-daemon", "-sock", "/sock"}, func(p Program) bool {
		return p.(Daemon).inner.SockPath == "/sock"
	}},
}

func isShowHelp(p Program) bool         { _, ok := p.(ShowHelp); return ok }
func isShowCorrectUsage(p Program) bool { _, ok := p.(ShowCorrectUsage); return ok }
func isShowVersion(p Program) bool      { _, ok := p.(ShowVersion); return ok }
func isShowBuildInfo(p Program) bool    { _, ok := p.(ShowBuildInfo); return ok }
func isDaemon(p Program) bool           { _, ok := p.(Daemon); return ok }
func isWeb(p Program) bool              { _, ok := p.(*web.Web); return ok }
func isShell(p Program) bool            { _, ok := p.(*shell.Shell); return ok }

func TestFindProgram(t *testing.T) {
	for i, test := range findProgramTests {
		f := parse(test.args)
		p := FindProgram(f)
		if !test.checker(p) {
			t.Errorf("test #%d (args = %q) failed", i, test.args)
		}
	}
}

func parse(args []string) *flagSet {
	f := newFlagSet()
	err := f.Parse(args)
	if err != nil {
		panic(fmt.Sprintln("bad flags in test", args))
	}
	return f
}

var programTests = []struct {
	run        []string
	stdin      string
	wantStdout string
	wantStderr string
}{
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

func TestPrograms(t *testing.T) {
	for _, test := range programTests {
		t.Run(strings.Join(test.run, " "), func(t *testing.T) {
			fd0, fd0w := mustMakePipe()
			fd1r, fd1 := mustMakePipe()
			fd2r, fd2 := mustMakePipe()

			fd0w.WriteString(test.stdin)
			fd0w.Close()

			Main([3]*os.File{fd0, fd1, fd2}, test.run)
			fd0.Close()
			fd1.Close()
			fd2.Close()

			if stdout := mustReadAllAndClose(fd1r); stdout != test.wantStdout {
				t.Errorf("Stdout mismatch")
				t.Logf("Got:\n%s", stdout)
				t.Logf("Want:\n%s", test.wantStdout)
			}
			if stderr := mustReadAllAndClose(fd2r); stderr != test.wantStderr {
				t.Errorf("Stderr mismatch")
				t.Logf("Got:\n%s", stderr)
				t.Logf("Want:\n%s", test.wantStderr)
			}
		})
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
