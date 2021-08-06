package shell

import (
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/prog"
	. "src.elv.sh/pkg/prog/progtest"
	. "src.elv.sh/pkg/testutil"
)

func TestShell_LegacyLibPath(t *testing.T) {
	f := setup(t)
	MustWriteFile(filepath.Join(f.home, ".elvish", "lib", "a.elv"), "echo mod a")

	exit := f.run(Elvish("-c", "use a"))
	TestExit(t, exit, 0)
	f.TestOut(t, 1, "mod a\n")
}

// Most high-level tests against Program are specific to either script mode or
// interactive mode, and are found in script_test.go and interact_test.go.

var incSHLVLTests = []struct {
	name    string
	old     string
	unset   bool
	wantNew string
}{
	{name: "normal", old: "10", wantNew: "11"},
	{name: "unset", unset: true, wantNew: "1"},
	{name: "invalid", old: "invalid", wantNew: "1"},
	// Other shells don't agree on what to do when SHLVL is negative:
	//
	// ~> E:SHLVL=-100 bash -c 'echo $SHLVL'
	// 0
	// ~> E:SHLVL=-100 zsh -c 'echo $SHLVL'
	// -99
	// ~> E:SHLVL=-100 fish -c 'echo $SHLVL'
	// 1
	//
	// Elvish follows Zsh here.
	{name: "negative", old: "-100", wantNew: "-99"},
}

func TestIncSHLVL(t *testing.T) {
	Setenv(t, env.SHLVL, "")

	for _, test := range incSHLVLTests {
		t.Run(test.name, func(t *testing.T) {
			if test.unset {
				os.Unsetenv(env.SHLVL)
			} else {
				os.Setenv(env.SHLVL, test.old)
			}

			restore := IncSHLVL()
			shlvl := os.Getenv(env.SHLVL)
			if shlvl != test.wantNew {
				t.Errorf("got SHLVL = %q, want %q", shlvl, test.wantNew)
			}

			restore()
			// Test that state of SHLVL is restored.
			restored, restoredSet := os.LookupEnv(env.SHLVL)
			if test.unset {
				if restoredSet {
					t.Errorf("SHLVL not unset")
				}
			} else {
				if restored != test.old {
					t.Errorf("SHLVL restored to %q, want %q", restored, test.old)
				}
			}
		})
	}
}

type fixture struct {
	*Fixture
	home string
}

func setup(t Cleanuper) fixture {
	Unsetenv(t, env.XDG_CONFIG_HOME)
	Unsetenv(t, env.XDG_DATA_HOME)
	home := TempHome(t)
	return fixture{Setup(t), home}
}

func (f fixture) run(args []string) int { return prog.Run(f.Fds(), args, Program{}) }
