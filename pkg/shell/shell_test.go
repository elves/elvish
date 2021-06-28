package shell

import (
	"os"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/testutil"
)

func TestShell_SHLVL_NormalCase(t *testing.T) {
	restore := testutil.WithTempEnv(env.SHLVL, "10")
	defer restore()

	testSHLVL(t, "11")
}

func TestShell_SHLVL_Unset(t *testing.T) {
	restore := testutil.WithTempEnv(env.SHLVL, "")
	defer restore()

	os.Unsetenv(env.SHLVL)

	testSHLVL(t, "1")
}

func TestShell_SHLVL_Invalid(t *testing.T) {
	restore := testutil.WithTempEnv(env.SHLVL, "invalid")
	defer restore()

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
	restore := testutil.WithTempEnv(env.SHLVL, "-100")
	defer restore()

	os.Setenv(env.SHLVL, "-100")
	testSHLVL(t, "-99")
}

func testSHLVL(t *testing.T, wantSHLVL string) {
	t.Helper()
	oldValue, oldOK := os.LookupEnv(env.SHLVL)

	restore := incSHLVL()
	shlvl := os.Getenv(env.SHLVL)
	if shlvl != wantSHLVL {
		t.Errorf("got SHLVL = %q, want %q", shlvl, wantSHLVL)
	}

	// Test that state of SHLVL is restored.
	restore()
	newValue, newOK := os.LookupEnv(env.SHLVL)
	if newValue != oldValue {
		t.Errorf("SHLVL not restored, %q -> %q", oldValue, newValue)
	}
	if oldOK != newOK {
		t.Errorf("SHLVL existence not restored, %v -> %v", oldOK, newOK)
	}
}
