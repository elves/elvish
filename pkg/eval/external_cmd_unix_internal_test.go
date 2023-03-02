//go:build unix

package eval

import (
	"os"
	"reflect"
	"syscall"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

func TestExec_Argv0Argv(t *testing.T) {
	dir := testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"bin": testutil.Dir{
			"elvish": testutil.File{Perm: 0755},
			"cat":    testutil.File{Perm: 0755},
		},
	})

	testutil.Setenv(t, "PATH", dir+"/bin")
	testutil.Setenv(t, env.SHLVL, "1")

	var tests = []struct {
		name      string
		code      string
		wantArgv0 string
		wantArgv  []string
		wantError bool
	}{
		{
			name:      "absolute path command",
			code:      "exec /bin/sh foo bar",
			wantArgv0: "/bin/sh",
			wantArgv:  []string{"/bin/sh", "foo", "bar"},
		},
		{
			name:      "relative path command",
			code:      "exec cat foo bar",
			wantArgv0: dir + "/bin/cat",
			wantArgv:  []string{dir + "/bin/cat", "foo", "bar"},
		},
		{
			name:      "no command",
			code:      "exec",
			wantArgv0: dir + "/bin/elvish",
			wantArgv:  []string{dir + "/bin/elvish"},
		},
		{
			name:      "bad command",
			code:      "exec bad",
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Helper()
			var (
				gotArgv0 string
				gotArgv  []string
			)
			syscallExec = func(argv0 string, argv []string, envv []string) error {
				gotArgv0 = argv0
				gotArgv = argv
				return nil
			}
			defer func() { syscallExec = syscall.Exec }()

			ev := NewEvaler()
			err := ev.Eval(parse.Source{Name: "[test]", Code: test.code}, EvalCfg{})

			if gotArgv0 != test.wantArgv0 {
				t.Errorf("got argv0 %q, want %q", gotArgv0, test.wantArgv0)
			}
			if !reflect.DeepEqual(gotArgv, test.wantArgv) {
				t.Errorf("got argv %q, want %q", gotArgv, test.wantArgv)
			}
			hasError := err != nil
			if hasError != test.wantError {
				t.Errorf("has error %v, want %v", hasError, test.wantError)
			}
		})
	}
}

func TestDecSHLVL(t *testing.T) {
	// Valid integers are decremented, regardless of sign
	testDecSHLVL(t, "-2", "-3")
	testDecSHLVL(t, "-1", "-2")
	testDecSHLVL(t, "0", "-1")
	testDecSHLVL(t, "1", "0")
	testDecSHLVL(t, "2", "1")
	testDecSHLVL(t, "3", "2")

	// Non-integers are kept unchanged
	testDecSHLVL(t, "", "")
	testDecSHLVL(t, "x", "x")
}

func testDecSHLVL(t *testing.T, oldValue, newValue string) {
	t.Helper()
	testutil.Setenv(t, env.SHLVL, oldValue)

	decSHLVL()
	if gotValue := os.Getenv(env.SHLVL); gotValue != newValue {
		t.Errorf("got new value %q, want %q", gotValue, newValue)
	}
}
