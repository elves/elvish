package shell

import (
	"os"
	"path/filepath"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/prog"
)

// RCPath returns the path of rc.elv, executed in interactive mode.
func RCPath() (string, error) {
	if legacyRC, exists := legacyDataPath("rc.elv", false); exists {
		return legacyRC, nil
	}
	return rcPath()
}

func LibPaths() ([]string, string, error) {
	paths, installPath, err := libPaths()
	if legacyLib, exists := legacyDataPath("lib", true); exists {
		paths = append(paths, legacyLib)
	}
	return paths, installPath, err
}

// Returns a SpawnConfig containing all the paths needed by the daemon. It
// respects overrides of sock and db from CLI flags.
func daemonPaths(flags *prog.Flags) (*daemondefs.SpawnConfig, error) {
	runDir, err := getSecureRunDir()
	if err != nil {
		return nil, err
	}
	sock := flags.Sock
	if sock == "" {
		sock = filepath.Join(runDir, "sock")
	}

	db := flags.DB
	if db == "" {
		if legacyPath, exists := legacyDataPath("db", false); exists {
			db = legacyPath
		} else {
			p, err := dbPath()
			if err != nil {
				return nil, err
			}
			db = p
		}
		err := os.MkdirAll(filepath.Dir(db), 0700)
		if err != nil {
			return nil, err
		}
	}
	return &daemondefs.SpawnConfig{DbPath: db, SockPath: sock, RunDir: runDir}, nil
}

// Returns a path in the legacy data directory path, and whether it exists and
// matches the expected file/directory property.
func legacyDataPath(name string, dir bool) (string, bool) {
	home, err := fsutil.GetHome("")
	if err != nil {
		return "", false
	}
	p := filepath.Join(home, ".elvish", name)
	info, err := os.Stat(p)
	if err != nil || info.IsDir() != dir {
		return "", false
	}
	return p, true
}
