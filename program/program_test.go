package program

import (
	"fmt"
	"testing"

	"github.com/elves/elvish/program/shell"
	"github.com/elves/elvish/program/web"
)

var findProgramTests = []struct {
	args    []string
	checker func(Program) bool
}{
	{[]string{"-help"}, isShowHelp},
	{[]string{"-version"}, isShowVersion},
	{[]string{"-buildinfo"}, isShowBuildInfo},
	{[]string{"-buildinfo", "-json"}, func(p Program) bool {
		return p.(ShowBuildInfo).JSON
	}},
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
	{[]string{"-daemon", "-forked", "1"}, func(p Program) bool {
		return p.(Daemon).inner.Forked == 1
	}},

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
