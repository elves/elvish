package shell

import (
	"os"
	"testing"

	. "github.com/elves/elvish/pkg/prog/progtest"
	"github.com/elves/elvish/pkg/util"
)

func TestShell_SHLVL_NormalCase(t *testing.T) {
	restore := saveEnv(util.EnvSHLVL)
	defer restore()

	os.Setenv(util.EnvSHLVL, "10")
	testSHLVL(t, "11")
}

func TestShell_SHLVL_Unset(t *testing.T) {
	restore := saveEnv(util.EnvSHLVL)
	defer restore()

	os.Unsetenv(util.EnvSHLVL)
	testSHLVL(t, "1")
}

func TestShell_SHLVL_Invalid(t *testing.T) {
	restore := saveEnv(util.EnvSHLVL)
	defer restore()

	os.Setenv(util.EnvSHLVL, "invalid")
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
	restore := saveEnv(util.EnvSHLVL)
	defer restore()

	os.Setenv(util.EnvSHLVL, "-100")
	testSHLVL(t, "-99")
}

func testSHLVL(t *testing.T, wantSHLVL string) {
	t.Helper()
	f := Setup()
	defer f.Cleanup()

	oldValue, oldOK := os.LookupEnv(util.EnvSHLVL)

	Script(f.Fds(), []string{"print $E:SHLVL"}, &ScriptConfig{Cmd: true})
	f.TestOut(t, 1, wantSHLVL)
	f.TestOut(t, 2, "")

	// Test that state of SHLVL is restored.
	newValue, newOK := os.LookupEnv(util.EnvSHLVL)
	if newValue != oldValue {
		t.Errorf("SHLVL not restored, %q -> %q", oldValue, newValue)
	}
	if oldOK != newOK {
		t.Errorf("SHLVL existence not restored, %v -> %v", oldOK, newOK)
	}
}
