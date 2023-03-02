//go:build unix

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/testutil"
)

var elvishDashUID = fmt.Sprintf("elvish-%d", os.Getuid())

func TestSecureRunDir_PrefersXDGWhenNeitherExists(t *testing.T) {
	xdg, _ := setupForSecureRunDir(t)
	testSecureRunDir(t, filepath.Join(xdg, "elvish"), false)
}

func TestSecureRunDir_PrefersXDGWhenBothExist(t *testing.T) {
	xdg, tmp := setupForSecureRunDir(t)

	os.MkdirAll(filepath.Join(xdg, "elvish"), 0700)
	os.MkdirAll(filepath.Join(tmp, elvishDashUID), 0700)

	testSecureRunDir(t, filepath.Join(xdg, "elvish"), false)
}

func TestSecureRunDir_PrefersTmpWhenOnlyItExists(t *testing.T) {
	_, tmp := setupForSecureRunDir(t)

	os.MkdirAll(filepath.Join(tmp, elvishDashUID), 0700)

	testSecureRunDir(t, filepath.Join(tmp, elvishDashUID), false)
}

func TestSecureRunDir_PrefersTmpWhenXdgEnvIsEmpty(t *testing.T) {
	_, tmp := setupForSecureRunDir(t)
	os.Setenv(env.XDG_RUNTIME_DIR, "")
	testSecureRunDir(t, filepath.Join(tmp, elvishDashUID), false)
}

func TestSecureRunDir_ReturnsErrorWhenUnableToMkdir(t *testing.T) {
	xdg, _ := setupForSecureRunDir(t)
	os.WriteFile(filepath.Join(xdg, "elvish"), nil, 0600)
	testSecureRunDir(t, "", true)
}

func setupForSecureRunDir(c testutil.Cleanuper) (xdgRuntimeDir, tmpDir string) {
	xdg := testutil.Setenv(c, env.XDG_RUNTIME_DIR, testutil.TempDir(c))
	tmp := testutil.Setenv(c, "TMPDIR", testutil.TempDir(c))
	return xdg, tmp
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
