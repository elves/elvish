package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/prog"
)

const legacyRcPathWarning = `Warning: ~/.elvish/rc.elv will be ignored from Elvish 0.20.0. Move it to its new location, as documented in https://elv.sh/ref/command.html#rc-file.`

func rcPath(w io.Writer) (string, error) {
	if legacyRC, exists := legacyDataPath("rc.elv", false); exists {
		fmt.Fprintln(w, legacyRcPathWarning)
		return legacyRC, nil
	}
	configHome, err := fsutil.ConfigHome()
	if err == nil {
		return filepath.Join(configHome, "rc.elv"), nil
	}
	return "", fmt.Errorf("find rc.elv: %w", err)
}

const legacyLibPathWarning = `Warning: ~/.elvish/lib will be ignored from Elvish 0.20.0. Move libraries to one of the new module search directories, as documented in https://elv.sh/ref/command.html#module-search-directories.`

func libPaths(w io.Writer) ([]string, error) {
	var paths []string

	configHome, err := fsutil.ConfigHome()
	if err != nil {
		return nil, fmt.Errorf("can't find roaming lib directory: %w", err)
	}
	paths = append(paths, filepath.Join(configHome, "lib"))

	dataHome, err := fsutil.DataHome()
	if err != nil {
		return nil, fmt.Errorf("can't find local lib directory: %w", err)
	}
	paths = append(paths, filepath.Join(dataHome, "lib"))

	if dataDirs := os.Getenv(env.XDG_DATA_DIRS); dataDirs != "" {
		// XDG requires the paths be joined with ":". However, on Windows ":"
		// appear after the drive letter, so it's infeasible to use it to also
		// join paths.
		for _, dataDir := range filepath.SplitList(dataDirs) {
			paths = append(paths, filepath.Join(dataDir, "elvish", "lib"))
		}
	} else {
		paths = append(paths, fsutil.DefaultDataDirs...)
	}

	if legacyLib, exists := legacyDataPath("lib", true); exists {
		fmt.Fprintln(w, legacyLibPathWarning)
		paths = append(paths, legacyLib)
	}
	return paths, nil
}

// Returns a SpawnConfig containing all the paths needed by the daemon. It
// respects overrides of sock and db from CLI flags.
func daemonPaths(p *prog.DaemonPaths, w io.Writer) (*daemondefs.SpawnConfig, error) {
	runDir, err := fsutil.SecureRunDir()
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
		db, err = dbPath(w)
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

const legacyDbPathWarning = `Warning: ~/.elvish/db will be ignored from Elvish 0.20.0. Kill the daemon with "use daemon; kill $daemon:pid", and move the db to its new location, as documented in https://elv.sh/ref/command.html#database-file. The daemon will respawn when you launch another Elvish instance.`

func dbPath(w io.Writer) (string, error) {
	if legacyDB, exists := legacyDataPath("db", false); exists {
		fmt.Fprintln(w, legacyDbPathWarning)
		return legacyDB, nil
	}

	stateHome, err := fsutil.StateHome()
	if err != nil {
		return "", fmt.Errorf("find db: %w", err)
	}
	return filepath.Join(stateHome, "db.bolt"), nil
}

// Returns a path in the legacy data directory path, and whether it exists and
// matches the expected file/directory property.
func legacyDataPath(name string, dir bool) (string, bool) {
	dataDir, exists := legacyDataDir()
	if !exists {
		return "", false
	}
	p := filepath.Join(dataDir, name)
	info, err := os.Stat(p)
	return p, err == nil && info.IsDir() == dir
}

// Returns the legacy data directory ~/.elvish and whether it exists as a
// directory.
func legacyDataDir() (string, bool) {
	home, err := fsutil.GetHome("")
	if err != nil {
		return "", false
	}
	p := filepath.Join(home, ".elvish")
	info, err := os.Stat(p)
	return p, err == nil && info.IsDir()
}
