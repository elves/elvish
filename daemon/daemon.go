// Package exec provides the entry point of the daemon sub-program and helpers
// to spawn a daemon process.
package daemon

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/util"
)

// Daemon keeps configurations for the daemon process.
type Daemon struct {
	Forked        int
	BinPath       string
	DbPath        string
	SockPath      string
	LogPathPrefix string
}

// Main is the entry point of the daemon sub-program.
func (d *Daemon) Main(serve func(string, string)) int {
	switch d.Forked {
	case 0:
		errored := false
		absify := func(f string, s *string) {
			if *s == "" {
				log.Println("flag", f, "is required for daemon")
				errored = true
				return
			}
			p, err := filepath.Abs(*s)
			if err != nil {
				log.Println("abs:", err)
				errored = true
			} else {
				*s = p
			}
		}
		absify("-bin", &d.BinPath)
		absify("-db", &d.DbPath)
		absify("-sock", &d.SockPath)
		absify("-logprefix", &d.LogPathPrefix)
		if errored {
			return 2
		}

		syscall.Umask(0077)
		return d.pseudoFork(
			&os.ProcAttr{
				// cd to /
				Dir: "/",
				// empty environment
				Env: nil,
				Sys: &syscall.SysProcAttr{Setsid: true},
			})
	case 1:
		return d.pseudoFork(&os.ProcAttr{})
	case 2:
		serve(d.SockPath, d.DbPath)
		return 0
	default:
		return 2
	}
}

// Spawn spawns a daemon in the background. It is supposed to be called from a
// client.
func (d *Daemon) Spawn() error {
	binPath := d.BinPath
	// Determine binPath.
	if binPath == "" {
		if len(os.Args) > 0 && path.IsAbs(os.Args[0]) {
			binPath = os.Args[0]
		} else {
			// Find elvish in PATH
			paths := strings.Split(os.Getenv("PATH"), ":")
			result, err := util.Search(paths, "elvish")
			if err != nil {
				return errors.New("cannot find elvish: " + err.Error())
			}
			binPath = result
		}
	}

	return forkExec(
		&os.ProcAttr{Files: []*os.File{nil, nil, os.Stderr}},
		0,
		binPath,
		d.DbPath,
		d.SockPath,
		d.LogPathPrefix,
		true,
	)
}

// pseudoFork forks a daemon. It is supposed to be called from the daemon.
func (d *Daemon) pseudoFork(attr *os.ProcAttr) int {
	err := forkExec(
		attr,
		d.Forked+1,
		d.BinPath,
		d.DbPath,
		d.SockPath,
		d.LogPathPrefix,
		false,
	)

	if err != nil {
		return 2
	}
	return 0
}

func forkExec(attr *os.ProcAttr, forkLevel int, binPath, dbPath, sockPath,
	logPathPrefix string, wait bool) error {
	p, err := os.StartProcess(binPath, []string{
		binPath,
		"-daemon",
		"-forked", strconv.Itoa(forkLevel),
		"-bin", binPath,
		"-db", dbPath,
		"-sock", sockPath,
		"-logprefix", logPathPrefix,
	}, attr)

	if err != nil {
		return err
	}
	if wait {
		_, err = p.Wait()
	} else {
		err = p.Release()
	}
	return err
}
