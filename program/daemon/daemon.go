// Package daemon provides the entry point of the daemon sub-program and helpers
// to spawn a daemon process.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Daemon keeps configurations for the daemon sub-program. It can be used both
// from the main function for running the daemon and from another process
// (typically the first Elvish shell session) for spawning a daemon.
type Daemon struct {
	// Forked is the number of times the daemon has forked itself. It is only
	// relevant in the daemon sub-program.
	Forked int
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

// Main is the entry point of the daemon sub-program. It takes a serve function
// and returns the exit status. The caller can call os.Exit after doing more
// cleanup work. The exact behavior of this function depends on the -forked
// flag:
//
// When d.Forked = 0 (which is how the daemon gets started), it forks another
// daemon process, with an empty environment, the working directory changed to
// /, with umask 0077, and in a new process session. All the path flags are
// resolved to absolute paths, and the -fork flag becomes 1.
//
// When d.Forked = 1, it simply forks another daemon process with all arguments
// passed as is, except -forked which becomes 2.
//
// When d.Forked = 2, it calls the serve function with the value of d.SockPath
// and d.DbPath.
//
// These 3 steps implement a standard demonizing procedure. The main deviation
// is that since Go does not support a raw fork call, it has to use fork-exec to
// simulate fork and pass all states via command line flags.
func (d *Daemon) Main(serve func(string, string)) error {
	switch d.Forked {
	case 0:
		var absifyError error
		absify := func(name string, path string) string {
			if absifyError != nil {
				return ""
			}
			if path == "" {
				absifyError = fmt.Errorf("flag %s is required for daemon", name)
				return ""
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				absifyError = fmt.Errorf("cannot convert %s to absolute path: %s", name, err)
			}
			return absPath
		}
		binPath := absify("-bin", d.BinPath)
		dbPath := absify("-db", d.DbPath)
		sockPath := absify("-sock", d.SockPath)
		logPathPrefix := absify("-logprefix", d.LogPathPrefix)
		if absifyError != nil {
			return absifyError
		}

		setUmask()
		return startProcess(
			binPath, 1,
			dbPath, sockPath, logPathPrefix,
			&os.ProcAttr{
				Dir: "/",        // cd to /
				Env: []string{}, // empty environment
				Sys: sysProAttrForFirstFork(),
			})
	case 1:
		return startProcess(
			d.BinPath, 2,
			d.DbPath, d.SockPath, d.LogPathPrefix, nil)
	case 2:
		serve(d.SockPath, d.DbPath)
		return nil
	default:
		return fmt.Errorf("-forked is %d, should be 0, 1 or 2", d.Forked)
	}
}

// Spawn spawns a daemon in the background by calling the Elvish binary with
// appropriate flags. If d.BinPath is not set, it attempts to derive the binary
// path using os.Executable(). The fields DbPath, SockPath and LogPathPrefix
// must be set.
//
// It is supposed to be called from a Elvish shell (as opposed to daemon) process.
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

	return startProcess(
		binPath, 0,
		d.DbPath, d.SockPath, d.LogPathPrefix, nil)
}

func startProcess(binPath string, forked int,
	dbPath, sockPath, logPathPrefix string, attr *os.ProcAttr) error {

	if attr == nil {
		attr = &os.ProcAttr{}
	}
	args := []string{
		binPath,
		"-daemon",
		"-forked", strconv.Itoa(forked),
		"-bin", binPath,
		"-db", dbPath,
		"-sock", sockPath,
		"-logprefix", logPathPrefix,
	}

	_, err := os.StartProcess(binPath, args, attr)
	return err
}
