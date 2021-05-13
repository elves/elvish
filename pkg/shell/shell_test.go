package shell

import (
	"os"
	"testing"

	"src.elv.sh/pkg/env"
	. "src.elv.sh/pkg/prog/progtest"
)

func TestShell_SHLVL_NormalCase(t *testing.T) {
	restore := saveEnv(env.Shlvl)
	defer restore()

	os.Setenv(env.Shlvl, "10")
	testSHLVL(t, "11")
}

func TestShell_SHLVL_Unset(t *testing.T) {
	restore := saveEnv(env.Shlvl)
	defer restore()

	os.Unsetenv(env.Shlvl)
	testSHLVL(t, "1")
}

func TestShell_SHLVL_Invalid(t *testing.T) {
	restore := saveEnv(env.Shlvl)
	defer restore()

	os.Setenv(env.Shlvl, "invalid")
	testSHLVL(t, "1")
}

func TestShell_NegativeSHLVL_Increments(t *testing.T) {
	// Other shells don't agree what the behavior should be:
	//
	// ~> E:SHLVL=-100 bash -c 'echo $SHLVL'
	// 0
	// ~> E:SHLVL=-100 zsh -c 'echo $SHLVL'
	// -99
	// ~> E:SHLVL=-100 fish -c 'echo $SHLVL'
	// 1
	//
	// Elvish follows Zsh here.
	restore := saveEnv(env.Shlvl)
	defer restore()

	os.Setenv(env.Shlvl, "-100")
	testSHLVL(t, "-99")
}

func testSHLVL(t *testing.T, wantSHLVL string) {
	t.Helper()
	f := Setup()
	defer f.Cleanup()

	oldValue, oldOK := os.LookupEnv(env.Shlvl)

	Script(f.Fds(), []string{"print $E:SHLVL"}, &ScriptConfig{Cmd: true})
	f.TestOut(t, 1, wantSHLVL)
	f.TestOut(t, 2, "")

	// Test that state of SHLVL is restored.
	newValue, newOK := os.LookupEnv(env.Shlvl)
	if newValue != oldValue {
		t.Errorf("SHLVL not restored, %q -> %q", oldValue, newValue)
	}
	if oldOK != newOK {
		t.Errorf("SHLVL existence not restored, %v -> %v", oldOK, newOK)
	}
}
