package shell

import (
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	. "src.elv.sh/pkg/prog/progtest"
	. "src.elv.sh/pkg/testutil"
)

func TestShell_LegacyLibPath(t *testing.T) {
	home := setupHomePaths(t)
	MustWriteFile(filepath.Join(home, ".elvish", "lib", "a.elv"), "echo mod a")

	Test(t, &Program{},
		ThatElvish("-c", "use a").WritesStdout("mod a\n"),
	)
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

// Common test utilities.

func setupHomePaths(t Cleanuper) string {
	Unsetenv(t, env.XDG_CONFIG_HOME)
	Unsetenv(t, env.XDG_DATA_HOME)
	return TempHome(t)
}
