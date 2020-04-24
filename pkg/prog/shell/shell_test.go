package shell

import (
	"os"
	"testing"

	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestShell_SHLVL_NormalCase(t *testing.T) {
	restore := saveEnv("SHLVL")
	defer restore()

	os.Setenv("SHLVL", "10")
	testSHLVL(t, "11")
}

func TestShell_SHLVL_Unset(t *testing.T) {
	restore := saveEnv("SHLVL")
	defer restore()

	os.Unsetenv("SHLVL")
	testSHLVL(t, "1")
}

func TestShell_SHLVL_Invalid(t *testing.T) {
	restore := saveEnv("SHLVL")
	defer restore()

	os.Setenv("SHLVL", "invalid")
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
	restore := saveEnv("SHLVL")
	defer restore()

	os.Setenv("SHLVL", "-100")
	testSHLVL(t, "-99")
}

func testSHLVL(t *testing.T, wantSHLVL string) {
	t.Helper()
	f := Setup()
	defer f.Cleanup()

	oldValue, oldOK := os.LookupEnv("SHLVL")

	Script(f.Fds(), []string{"print $E:SHLVL"}, &ScriptConfig{Cmd: true})
	f.TestOut(t, 1, wantSHLVL)

	// Test that state of SHLVL is restored.
	newValue, newOK := os.LookupEnv("SHLVL")
	if newValue != oldValue {
		t.Errorf("SHLVL not restored, %q -> %q", oldValue, newValue)
	}
	if oldOK != newOK {
		t.Errorf("SHLVL existence not restored, %v -> %v", oldOK, newOK)
	}
}
