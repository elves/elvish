package shell

import (
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/must"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestShell_LibPath_XDGPaths(t *testing.T) {
	xdgConfigHome := testutil.TempDir(t)
	testutil.ApplyDirIn(testutil.Dir{
		"elvish": testutil.Dir{
			"lib": testutil.Dir{
				"a.elv": "echo a from xdg-config-home",
			},
		},
	}, xdgConfigHome)
	testutil.Setenv(t, env.XDG_CONFIG_HOME, xdgConfigHome)

	xdgDataHome := testutil.TempDir(t)
	testutil.ApplyDirIn(testutil.Dir{
		"elvish": testutil.Dir{
			"lib": testutil.Dir{
				"a.elv": "echo a from xdg-data-home",
				"b.elv": "echo b from xdg-data-home",
			},
		},
	}, xdgDataHome)
	testutil.Setenv(t, env.XDG_DATA_HOME, xdgDataHome)

	xdgDataDir1 := testutil.TempDir(t)
	testutil.ApplyDirIn(testutil.Dir{
		"elvish": testutil.Dir{
			"lib": testutil.Dir{
				"a.elv": "echo a from xdg-data-dir-1",
				"b.elv": "echo b from xdg-data-dir-1",
				"c.elv": "echo c from xdg-data-dir-1",
			},
		},
	}, xdgDataDir1)
	xdgDataDir2 := testutil.TempDir(t)
	testutil.ApplyDirIn(testutil.Dir{
		"elvish": testutil.Dir{
			"lib": testutil.Dir{
				"a.elv": "echo a from xdg-data-dir-2",
				"b.elv": "echo b from xdg-data-dir-2",
				"c.elv": "echo c from xdg-data-dir-2",
				"d.elv": "echo d from xdg-data-dir-2",
			},
		},
	}, xdgDataDir2)
	testutil.Setenv(t, env.XDG_DATA_DIRS,
		xdgDataDir1+string(filepath.ListSeparator)+xdgDataDir2)

	Test(t, &Program{},
		ThatElvish("-c", "use a").WritesStdout("a from xdg-config-home\n"),
		ThatElvish("-c", "use b").WritesStdout("b from xdg-data-home\n"),
		ThatElvish("-c", "use c").WritesStdout("c from xdg-data-dir-1\n"),
		ThatElvish("-c", "use d").WritesStdout("d from xdg-data-dir-2\n"),
	)
}

func TestShell_LibPath_Legacy(t *testing.T) {
	home := setupCleanHomePaths(t)
	must.WriteFile(filepath.Join(home, ".elvish", "lib", "a.elv"), "echo mod a")

	Test(t, &Program{},
		ThatElvish("-c", "use a").
			WritesStdout("mod a\n").
			WritesStderrContaining(legacyLibPathWarning),
	)
}

// Most high-level tests against Program are specific to either script mode or
// interactive mode, and are found in script_test.go and interact_test.go.

var noColorTests = []struct {
	name       string
	value      string
	unset      bool
	wantRedFoo string
}{
	{name: "unset", unset: true, wantRedFoo: "\033[;31mfoo\033[m"},
	{name: "empty", value: "", wantRedFoo: "\033[;31mfoo\033[m"},
	{name: "non-empty", value: "yes", wantRedFoo: "\033[mfoo"},
}

func TestShell_NO_COLOR(t *testing.T) {
	for _, test := range noColorTests {
		t.Run(test.name, func(t *testing.T) {
			setOrUnsetenv(t, env.NO_COLOR, test.unset, test.value)
			Test(t, &Program{},
				ThatElvish("-c", "print (styled foo red)").
					WritesStdout(test.wantRedFoo))
		})
	}
}

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

func TestShell_SHLVL(t *testing.T) {
	for _, test := range incSHLVLTests {
		t.Run(test.name, func(t *testing.T) {
			setOrUnsetenv(t, env.SHLVL, test.unset, test.old)
			Test(t, &Program{},
				ThatElvish("-c", "print $E:SHLVL").WritesStdout(test.wantNew))

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

func setOrUnsetenv(t *testing.T, name string, unset bool, set string) {
	if unset {
		testutil.Unsetenv(t, name)
	} else {
		testutil.Setenv(t, name, set)
	}
}

// Common test utilities.

func setupCleanHomePaths(t testutil.Cleanuper) string {
	testutil.Unsetenv(t, env.XDG_CONFIG_HOME)
	testutil.Unsetenv(t, env.XDG_DATA_HOME)
	return testutil.TempHome(t)
}
