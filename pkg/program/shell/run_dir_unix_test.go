// +build !windows,!plan9

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestGetSecureRunDir(t *testing.T) {
	xdgRuntimeDir, xdgCleanup := util.TestDir()
	defer xdgCleanup()

	tmpDir, tmpCleanup := util.TestDir()
	defer tmpCleanup()

	uid := os.Getuid()

	tests := []struct {
		name          string
		xdgRuntimeDir string
		tmpdir        string
		want          string
	}{
		{
			name:          "prefer XDG_RUNTIME_DIR over TMPDIR",
			xdgRuntimeDir: xdgRuntimeDir,
			tmpdir:        tmpDir,
			want:          filepath.Join(xdgRuntimeDir, "elvish"),
		},
		{
			name:          "use XDG_RUNTIME_DIR when TMPDIR is not set",
			xdgRuntimeDir: xdgRuntimeDir,
			tmpdir:        "",
			want:          filepath.Join(xdgRuntimeDir, "elvish"),
		},
		{
			name:          "fallback to TMPDIR when XDG_RUNTIME_DIR is not set",
			xdgRuntimeDir: "",
			tmpdir:        tmpDir,
			want:          filepath.Join(tmpDir, fmt.Sprintf("elvish-%d", uid)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Helper()

			envOverrides := map[string]string{
				"XDG_RUNTIME_DIR": test.xdgRuntimeDir,
				"TMPDIR":          test.tmpdir,
			}

			restore := withTempEnvs(envOverrides)
			defer restore()

			runDir, err := getSecureRunDir()
			if err != nil {
				t.Errorf("Could not get secure run dir: %v", err)
			}

			wantRunDir := test.want
			if runDir != wantRunDir {
				t.Errorf("Got run dir %v, want %v", runDir, wantRunDir)
			}

			info, err := os.Stat(runDir)
			if err != nil {
				t.Errorf("Could not stat run dir: %v", err)
			}

			stat := info.Sys().(*syscall.Stat_t)
			if int(stat.Uid) != uid {
				t.Errorf("Invalid run dir owner")
			}
			if stat.Mode&077 != 0 {
				t.Errorf("Invalid run dir permissions")
			}
		})
	}
}

// TODO: Move to the util package and add tests
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
