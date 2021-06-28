package shell

import (
	"os"
	"path/filepath"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/prog"
)

// Paths keeps all paths required for the Elvish runtime.
type Paths struct {
	DataDir string
	Rc      string
	LibDir  string
}

// DataPaths returns all the data paths needed by the shell.
func DataPaths() (Paths, error) {
	dataDir, err := ensureDataDir()
	if err != nil {
		return Paths{}, err
	}
	return Paths{
		DataDir: dataDir,
		Rc:      filepath.Join(dataDir, "rc.elv"),
		LibDir:  filepath.Join(dataDir, "lib"),
	}, nil
}

// Ensures Elvish's data directory exists, creating it if necessary. It returns
// the path to the data directory (never with a trailing slash) and possible
// error.
func ensureDataDir() (string, error) {
	home, err := fsutil.GetHome("")
	if err != nil {
		return "", err
	}
	ddir := home + "/.elvish"
	return ddir, os.MkdirAll(ddir, 0700)
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
		dataDir, err := ensureDataDir()
		if err != nil {
			return nil, err
		}
		db = filepath.Join(dataDir, "db")
	}
	return &daemondefs.SpawnConfig{DbPath: db, SockPath: sock, RunDir: runDir}, nil
}
