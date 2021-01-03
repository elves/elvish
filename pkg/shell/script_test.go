package shell

import (
	"testing"

	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestScript_File(t *testing.T) {
	f := Setup()
	defer f.Cleanup()
	MustWriteFile("a.elv", "echo hello")

	Script(f.Fds(), []string{"a.elv"}, &ScriptConfig{})

	f.TestOut(t, 1, "hello\n")
	f.TestOut(t, 2, "")
}

func TestScript_BadFile(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	ret := Script(f.Fds(), []string{"a.elv"}, &ScriptConfig{})

	if ret != 2 {
		t.Errorf("got ret %v, want 2", ret)
	}
	f.TestOutSnippet(t, 2, "cannot read script")
	f.TestOut(t, 1, "")
}

func TestScript_Cmd(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

	Script(f.Fds(), []string{"echo hello"}, &ScriptConfig{Cmd: true})

	f.TestOut(t, 1, "hello\n")
	f.TestOut(t, 2, "")
}

var scriptErrorTests = []struct {
	name        string
	code        string
	compileOnly bool
	json        bool
	wantExit    int
	wantOut     string
	wantErr     string
}{
	{name: "parse error",
		code:     "echo [",
		wantExit: 2,
		wantErr:  "parse error"},
	{name: "parse error with -compileonly and -json",
		code: "echo [", compileOnly: true, json: true,
		wantExit: 2,
		wantOut:  `[{"fileName":"code from -c","start":6,"end":6,"message":"should be ']'"}]` + "\n"},
	{name: "multiple parse errors with -compileonly and -json",
		code: "echo [{", compileOnly: true, json: true,
		wantExit: 2,
		wantOut:  `[{"fileName":"code from -c","start":7,"end":7,"message":"should be ',' or '}'"},{"fileName":"code from -c","start":7,"end":7,"message":"should be ']'"}]` + "\n"},
	{name: "compile error",
		code:     "echo $a",
		wantExit: 2,
		wantErr:  "compilation error"},
	{name: "compile error with -compileonly and -json",
		code: "echo $a", compileOnly: true, json: true,
		wantExit: 2,
		wantOut:  `[{"fileName":"code from -c","start":5,"end":7,"message":"variable $a not found"}]` + "\n"},
	{name: "parse error and compile error with -compileonly and -json",
		code: "echo [$a", compileOnly: true, json: true,
		wantExit: 2,
		wantOut:  `[{"fileName":"code from -c","start":8,"end":8,"message":"should be ']'"},{"fileName":"code from -c","start":6,"end":8,"message":"variable $a not found"}]` + "\n"},
	{name: "exception",
		code:     "fail failure",
		wantExit: 2,
		wantOut:  "",
		wantErr:  "fail failure"},
	{name: "exception with -compileonly",
		code: "fail failure", compileOnly: true,
		wantExit: 0},
}

func TestScript_Error(t *testing.T) {
	for _, test := range scriptErrorTests {
		t.Run(test.name, func(t *testing.T) {
			f := Setup()
			defer f.Cleanup()
			exit := Script(f.Fds(), []string{test.code}, &ScriptConfig{
				Cmd: true, CompileOnly: test.compileOnly, JSON: test.json})
			if exit != test.wantExit {
				t.Errorf("got exit code %v, want 2", test.wantExit)
			}
			f.TestOut(t, 1, test.wantOut)
			// When testing stderr output, we either only test that there is no
			// output at all, or that the output contains a string; we never
			// test it in full, as it is intended for human consumption and may
			// change.
			if test.wantErr == "" {
				f.TestOut(t, 2, "")
			} else {
				f.TestOutSnippet(t, 2, test.wantErr)
			}
		})
	}
}
