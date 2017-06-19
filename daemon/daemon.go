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
	Forked   int
	BinPath  string
	DbPath   string
	SockPath string
	LogPath  string
}

// closeFd is used in syscall.ProcAttr.Files to signify closing a fd.
const closeFd = ^uintptr(0)

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
		absify("-log", &d.LogPath)
		if errored {
			return 2
		}

		syscall.Umask(0077)
		return d.pseudoFork(
			&syscall.ProcAttr{
				// cd to /
				Dir: "/",
				// empty environment
				Env: nil,
				// inherit stderr only for logging
				Files: []uintptr{closeFd, closeFd, 2},
				Sys:   &syscall.SysProcAttr{Setsid: true},
			})
	case 1:
		return d.pseudoFork(
			&syscall.ProcAttr{
				Files: []uintptr{closeFd, closeFd, 2},
			})
	case 2:
		serve(d.SockPath, d.DbPath)
		panic("unreachable")
	default:
		return 2
	}
}

// Spawn spawns a daemon in the background. It is supposed to be called from a
// client.
func (d *Daemon) Spawn(logpath string) error {
	// Determine binpath.
	if d.BinPath == "" {
		if len(os.Args) > 0 && path.IsAbs(os.Args[0]) {
			d.BinPath = os.Args[0]
		} else {
			// Find elvish in PATH
			paths := strings.Split(os.Getenv("PATH"), ":")
			binpath, err := util.Search(paths, "elvish")
			if err != nil {
				return errors.New("cannot find elvish: " + err.Error())
			}
			d.BinPath = binpath
		}
	}
	d.LogPath = logpath
	return d.forkExec(nil, 0)
}

// pseudoFork forks a daemon. It is supposed to be called from the daemon.
func (d *Daemon) pseudoFork(attr *syscall.ProcAttr) int {
	err := d.forkExec(attr, d.Forked+1)
	if err != nil {
		return 2
	}
	return 0
}

func (d *Daemon) forkExec(attr *syscall.ProcAttr, forklevel int) error {
	_, err := syscall.ForkExec(d.BinPath, []string{
		d.BinPath,
		"-daemon",
		"-forked", strconv.Itoa(forklevel),
		"-bin", d.BinPath,
		"-db", d.DbPath,
		"-sock", d.SockPath,
		"-log", d.LogPath,
	}, attr)
	return err
}
