// +build !windows,!plan9

package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/elves/elvish/pkg/env"
	"github.com/elves/elvish/pkg/testutil"
)

// TODO(xiaq): Rewrite these tests to test the exported MakePaths instead of the
// unexported getSecureRunDir.

var elvishDashUID = fmt.Sprintf("elvish-%d", os.Getuid())

func TestGetSecureRunDir_PrefersXDGWhenNeitherExists(t *testing.T) {
	xdg, _, cleanup := setupForSecureRunDir()
	defer cleanup()
	testSecureRunDir(t, filepath.Join(xdg, "elvish"), false)
}

func TestGetSecureRunDir_PrefersXDGWhenBothExist(t *testing.T) {
	xdg, tmp, cleanup := setupForSecureRunDir()
	defer cleanup()

	os.MkdirAll(filepath.Join(xdg, "elvish"), 0700)
	os.MkdirAll(filepath.Join(tmp, elvishDashUID), 0700)

	testSecureRunDir(t, filepath.Join(xdg, "elvish"), false)
}

func TestGetSecureRunDir_PrefersTmpWhenOnlyItExists(t *testing.T) {
	_, tmp, cleanup := setupForSecureRunDir()
	defer cleanup()

	os.MkdirAll(filepath.Join(tmp, elvishDashUID), 0700)

	testSecureRunDir(t, filepath.Join(tmp, elvishDashUID), false)
}

func TestGetSecureRunDir_PrefersTmpWhenXdgEnvIsEmpty(t *testing.T) {
	_, tmp, cleanup := setupForSecureRunDir()
	defer cleanup()
	os.Setenv(env.XDG_RUNTIME_DIR, "")
	testSecureRunDir(t, filepath.Join(tmp, elvishDashUID), false)
}

func TestGetSecureRunDir_ReturnsErrorWhenUnableToMkdir(t *testing.T) {
	xdg, _, cleanup := setupForSecureRunDir()
	defer cleanup()
	ioutil.WriteFile(filepath.Join(xdg, "elvish"), nil, 0600)
	testSecureRunDir(t, "", true)
}

func setupForSecureRunDir() (xdgRuntimeDir, tmpDir string, cleanup func()) {
	xdgRuntimeDir, xdgCleanup := testutil.TestDir()
	tmpDir, tmpCleanup := testutil.TestDir()
	envCleanup := withTempEnvs(map[string]string{
		env.XDG_RUNTIME_DIR: xdgRuntimeDir,
		"TMPDIR":            tmpDir,
	})
	return xdgRuntimeDir, tmpDir, func() {
		envCleanup()
		tmpCleanup()
		xdgCleanup()
	}
}

func testSecureRunDir(t *testing.T, wantRunDir string, wantErr bool) {
	runDir, err := getSecureRunDir()
	if runDir != wantRunDir {
		t.Errorf("got rundir %q, want %q", runDir, wantRunDir)
	}
	if wantErr && err == nil {
		t.Errorf("got nil err, want non-nil")
	} else if !wantErr && err != nil {
		t.Errorf("got err %v, want nil err", err)
	}
}

// TODO(xiaq): Move to the testutil package and add tests.
func withTempEnvs(envOverrides map[string]string) func() {
	valuesToRestore := map[string]string{}

	for key, value := range envOverrides {
		original, exists := os.LookupEnv(key)

		os.Setenv(key, value)

		if exists {
			valuesToRestore[key] = original
		}
	}

	return func() {
		for key := range envOverrides {
			value, exists := valuesToRestore[key]
			if exists {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}
