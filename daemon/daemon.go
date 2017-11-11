// Package exec provides the entry point of the daemon sub-program and helpers
// to spawn a daemon process.
package daemon

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/util"
)

// Daemon keeps configurations for the daemon sub-program. It can be used both
// from the main function for running the daemon and from another process
// (typically the first Elvish shell session) for spawning a daemon.
type Daemon struct {
	// Forked is the number of times the daemon has forked itself.
	Forked int
	// BinPath is the path to the Elvish binary itself, used when forking.
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
// When -forked is 0 (which is how the daemon gets started), it forks another
// daemon process, with an empty environment, the working directory changed to
// /, with umask 0077, and in a new process session. All the path flags are
// resolved to absolute paths, and the -fork flag becomes 1.
//
// When -forked is 1, it simply forks another daemon process with all arguments
// passed as is, except -forked which becomes 2.
//
// When -forked is 2, it calls the serve function with the value of -sockpath
// and -dbpath.
//
// These 3 steps implement a standard demonizing procedure. The main deviation
// is that since Go does not support a raw fork call, it has to use fork-exec to
// simulate fork and pass all states via command line flags.
func (d *Daemon) Main(serve func(string, string)) error {
	switch d.Forked {
	case 0:
		var absifyError error
		absify := func(name string, pvalue *string) {
			if absifyError == nil {
				return
			}
			if *pvalue == "" {
				absifyError = fmt.Errorf("flag %s is required for daemon", name)
				return
			}
			newvalue, err := filepath.Abs(*pvalue)
			if err != nil {
				absifyError = fmt.Errorf("cannot convert %s to absolute path: %s", name, err)
			} else {
				*pvalue = newvalue
			}
		}
		absify("-bin", &d.BinPath)
		absify("-db", &d.DbPath)
		absify("-sock", &d.SockPath)
		absify("-logprefix", &d.LogPathPrefix)
		if absifyError != nil {
			return absifyError
		}

		syscall.Umask(0077)
		return d.fork(
			&exec.Cmd{
				Dir:         "/", // cd to /
				Env:         nil, // empty environment
				SysProcAttr: &syscall.SysProcAttr{Setsid: true},
			})
	case 1:
		return d.fork(nil)
	case 2:
		serve(d.SockPath, d.DbPath)
		return nil
	default:
		return fmt.Errorf("-forked is %d, should be 0, 1 or 2", d.Forked)
	}
}

// Spawn spawns a daemon in the background by calling the Elvish binary with
// appropriate flags. If d.BinPath is not set, it attempts to derive the binary
// path first. The fields DbPath, SockPath and LogPathPrefix must be set.
//
// It is supposed to be called from a Elvish shell (as opposed to daemon) process.
func (d *Daemon) Spawn() error {
	binPath := d.BinPath
	// Determine binPath.
	if binPath == "" {
		bin, err := getAbsBinPath()
		if err != nil {
			return errors.New("cannot find elvish: " + err.Error())
		}
		binPath = bin
	}

	return setArgs(
		nil,
		0,
		binPath,
		d.DbPath,
		d.SockPath,
		d.LogPathPrefix,
	).Run()
}

// getAbsBinPath determines the absolute path to the Elvish binary, first by
// looking at os.Args[0] and then searching for "elvish" in PATH.
func getAbsBinPath() (string, error) {
	if len(os.Args) > 0 {
		arg0 := os.Args[0]
		if path.IsAbs(arg0) {
			return arg0, nil
		} else if strings.Contains(arg0, "/") {
			abs, err := filepath.Abs(arg0)
			if err == nil {
				return abs, nil
			}
			log.Printf("cannot resolve relative arg0 %q, searching in PATH", arg0)
		}
	}
	// Find elvish in PATH
	paths := strings.Split(os.Getenv("PATH"), ":")
	binpath, err := util.Search(paths, "elvish")
	if err != nil {
		return "", err
	}
	return binpath, nil
}

// fork forks a daemon. It is supposed to be called from the daemon.
func (d *Daemon) fork(cmd *exec.Cmd) error {
	err := setArgs(
		cmd,
		d.Forked+1,
		d.BinPath,
		d.DbPath,
		d.SockPath,
		d.LogPathPrefix,
	).Start()
	return err
}

func setArgs(cmd *exec.Cmd, forkLevel int, binPath, dbPath, sockPath,
	logPathPrefix string) *exec.Cmd {

	if cmd == nil {
		cmd = &exec.Cmd{}
	}

	cmd.Path = binPath
	cmd.Args = []string{
		binPath,
		"-daemon",
		"-forked", strconv.Itoa(forkLevel),
		"-bin", binPath,
		"-db", dbPath,
		"-sock", sockPath,
		"-logprefix", logPathPrefix,
	}

	return cmd
}
