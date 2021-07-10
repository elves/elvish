// +build !windows,!plan9

package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/testutil"
)

var elvishDashUID = fmt.Sprintf("elvish-%d", os.Getuid())

func TestSecureRunDir_PrefersXDGWhenNeitherExists(t *testing.T) {
	xdg, _, cleanup := setupForSecureRunDir()
	defer cleanup()
	testSecureRunDir(t, filepath.Join(xdg, "elvish"), false)
}

func TestSecureRunDir_PrefersXDGWhenBothExist(t *testing.T) {
	xdg, tmp, cleanup := setupForSecureRunDir()
	defer cleanup()

	os.MkdirAll(filepath.Join(xdg, "elvish"), 0700)
	os.MkdirAll(filepath.Join(tmp, elvishDashUID), 0700)

	testSecureRunDir(t, filepath.Join(xdg, "elvish"), false)
}

func TestSecureRunDir_PrefersTmpWhenOnlyItExists(t *testing.T) {
	_, tmp, cleanup := setupForSecureRunDir()
	defer cleanup()

	os.MkdirAll(filepath.Join(tmp, elvishDashUID), 0700)

	testSecureRunDir(t, filepath.Join(tmp, elvishDashUID), false)
}

func TestSecureRunDir_PrefersTmpWhenXdgEnvIsEmpty(t *testing.T) {
	_, tmp, cleanup := setupForSecureRunDir()
	defer cleanup()
	os.Setenv(env.XDG_RUNTIME_DIR, "")
	testSecureRunDir(t, filepath.Join(tmp, elvishDashUID), false)
}

func TestSecureRunDir_ReturnsErrorWhenUnableToMkdir(t *testing.T) {
	xdg, _, cleanup := setupForSecureRunDir()
	defer cleanup()
	ioutil.WriteFile(filepath.Join(xdg, "elvish"), nil, 0600)
	testSecureRunDir(t, "", true)
}

func setupForSecureRunDir() (xdgRuntimeDir, tmpDir string, cleanup func()) {
	xdgRuntimeDir, xdgCleanup := testutil.TestDir()
	tmpDir, tmpCleanup := testutil.TestDir()

	restore1 := testutil.WithTempEnv(env.XDG_RUNTIME_DIR, xdgRuntimeDir)
	restore2 := testutil.WithTempEnv("TMPDIR", tmpDir)

	return xdgRuntimeDir, tmpDir, func() {
		restore2()
		restore1()
		tmpCleanup()
		xdgCleanup()
	}
}

func testSecureRunDir(t *testing.T, wantRunDir string, wantErr bool) {
	runDir, err := secureRunDir()
	if runDir != wantRunDir {
		t.Errorf("got rundir %q, want %q", runDir, wantRunDir)
	}
	if wantErr && err == nil {
		t.Errorf("got nil err, want non-nil")
	} else if !wantErr && err != nil {
		t.Errorf("got err %v, want nil err", err)
	}
}
