package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SpawnConfig keeps configurations for spawning the daemon.
type SpawnConfig struct {
	// BinPath is the path to the Elvish binary itself, used when forking. This
	// field is used only when spawning the daemon. If empty, it is
	// automatically determined with os.Executable.
	BinPath string
	// DbPath is the path to the database.
	DbPath string
	// SockPath is the path to the socket on which the daemon will serve
	// requests.
	SockPath string
	// LogPathPrefix is used to derive the name of the log file by adding the
	// pid.
	LogPathPrefix string
}

// Spawn spawns a daemon process in the background by invoking BinPath, passing
// DbPath, SockPath and LogPathPrefix as command-line arguments after resolving
// them to absolute paths. A suitable ProcAttr is chosen depending on the OS and
// makes sure that the daemon is detached from the current terminal (so that it
// is not affected by I/O or signals in the current terminal), and keeps running
// after the current process quits.
func Spawn(cfg *SpawnConfig) error {
	binPath := cfg.BinPath
	// Determine binPath.
	if binPath == "" {
		bin, err := os.Executable()
		if err != nil {
			return errors.New("cannot find elvish: " + err.Error())
		}
		binPath = bin
	}

	var pathError error
	abs := func(name string, path string) string {
		if pathError != nil {
			return ""
		}
		if path == "" {
			pathError = fmt.Errorf("%s is required for spawning daemon", name)
			return ""
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			pathError = fmt.Errorf("cannot resolve %s to absolute path: %s", name, err)
		}
		return absPath
	}
	binPath = abs("BinPath", binPath)
	dbPath := abs("DbPath", cfg.DbPath)
	sockPath := abs("SockPath", cfg.SockPath)
	logPathPrefix := abs("LogPathPrefix", cfg.LogPathPrefix)
	if pathError != nil {
		return pathError
	}

	args := []string{
		binPath,
		"-daemon",
		"-bin", binPath,
		"-db", dbPath,
		"-sock", sockPath,
		"-logprefix", logPathPrefix,
	}

	// TODO Redirect daemon stdout and stderr

	_, err := os.StartProcess(binPath, args, procAttrForSpawn())
	return err
}
