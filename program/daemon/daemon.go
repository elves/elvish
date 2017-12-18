// Package daemon provides the entry point of the daemon sub-program and helpers
// to spawn a daemon process.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Daemon keeps configurations for the daemon sub-program. It can be used both
// from the main function for running the daemon and from another process
// (typically the first Elvish shell session) for spawning a daemon.
type Daemon struct {
	// BinPath is the path to the Elvish binary itself, used when forking. This
	// field is optional only when spawning the daemon.
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

// Main is the entry point of the daemon sub-program. It simply sets the umask
// (if relevant) and runs serve. It always return a nil error, since any errors
// encountered is logged in the serve function.
func (d *Daemon) Main(serve func(string, string)) error {
	setUmask()
	serve(d.SockPath, d.DbPath)
	return nil
}

// Spawn spawns a daemon process in the background by invoking BinPath, passing
// DbPath, SockPath and LogPathPrefix as command-line arguments after resolving
// them to absolute paths. A suitable ProcAttr is chosen depending on the OS and
// makes sure that the daemon is detached from the current terminal (so that it
// is not affected by I/O or signals in the current terminal), and keeps running
// after the current process quits.
func (d *Daemon) Spawn() error {
	binPath := d.BinPath
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
	dbPath := abs("DbPath", d.DbPath)
	sockPath := abs("SockPath", d.SockPath)
	logPathPrefix := abs("LogPathPrefix", d.LogPathPrefix)
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
