package shell

import (
	"os"
	"path/filepath"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/prog"
)

func rcPath() (string, error) {
	if legacyRC, exists := legacyDataPath("rc.elv", false); exists {
		return legacyRC, nil
	}
	return newRCPath()
}

func libPaths() ([]string, error) {
	paths, err := newLibPaths()
	if legacyLib, exists := legacyDataPath("lib", true); exists {
		paths = append(paths, legacyLib)
	}
	return paths, err
}

// Returns a SpawnConfig containing all the paths needed by the daemon. It
// respects overrides of sock and db from CLI flags.
func daemonPaths(p *prog.DaemonPaths) (*daemondefs.SpawnConfig, error) {
	runDir, err := secureRunDir()
	if err != nil {
		return nil, err
	}
	sock := p.Sock
	if sock == "" {
		sock = filepath.Join(runDir, "sock")
	}

	db := p.DB
	if db == "" {
		var err error
		db, err = dbPath()
		if err != nil {
			return nil, err
		}
		err = os.MkdirAll(filepath.Dir(db), 0700)
		if err != nil {
			return nil, err
		}
	}
	return &daemondefs.SpawnConfig{DbPath: db, SockPath: sock, RunDir: runDir}, nil
}

func dbPath() (string, error) {
	if legacyDB, exists := legacyDataPath("db", false); exists {
		return legacyDB, nil
	}
	return newDBPath()
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
