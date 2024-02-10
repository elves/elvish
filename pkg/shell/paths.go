package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/prog"
)

func rcPath() (string, error) {
	if configHome := os.Getenv(env.XDG_CONFIG_HOME); configHome != "" {
		return filepath.Join(configHome, "elvish", "rc.elv"), nil
	} else if configHome, err := defaultConfigHome(); err == nil {
		return filepath.Join(configHome, "elvish", "rc.elv"), nil
	} else {
		return "", fmt.Errorf("find rc.elv: %w", err)
	}
}

func libPaths() ([]string, error) {
	var paths []string

	if configHome := os.Getenv(env.XDG_CONFIG_HOME); configHome != "" {
		paths = append(paths, filepath.Join(configHome, "elvish", "lib"))
	} else if configHome, err := defaultConfigHome(); err == nil {
		paths = append(paths, filepath.Join(configHome, "elvish", "lib"))
	} else {
		return nil, fmt.Errorf("find roaming lib directory: %w", err)
	}

	if dataHome := os.Getenv(env.XDG_DATA_HOME); dataHome != "" {
		paths = append(paths, filepath.Join(dataHome, "elvish", "lib"))
	} else if dataHome, err := defaultDataHome(); err == nil {
		paths = append(paths, filepath.Join(dataHome, "elvish", "lib"))
	} else {
		return nil, fmt.Errorf("find local lib directory: %w", err)
	}

	if dataDirs := os.Getenv(env.XDG_DATA_DIRS); dataDirs != "" {
		// XDG requires the paths be joined with ":". However, on Windows ":"
		// appear after the drive letter, so it's infeasible to use it to also
		// join paths.
		for _, dataDir := range filepath.SplitList(dataDirs) {
			paths = append(paths, filepath.Join(dataDir, "elvish", "lib"))
		}
	} else {
		paths = append(paths, defaultDataDirs...)
	}

	return paths, nil
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
	if stateHome := os.Getenv(env.XDG_STATE_HOME); stateHome != "" {
		return filepath.Join(stateHome, "elvish", "db.bolt"), nil
	} else if stateHome, err := defaultStateHome(); err == nil {
		return filepath.Join(stateHome, "elvish", "db.bolt"), nil
	} else {
		return "", fmt.Errorf("find db: %w", err)
	}
}
