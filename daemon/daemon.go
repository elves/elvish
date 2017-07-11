// Package exec provides the entry point of the daemon sub-program and helpers
// to spawn a daemon process.
package daemon

import (
	"errors"
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
			&exec.Cmd{
				Dir:         "/", // cd to /
				Env:         nil, // empty environment
				SysProcAttr: &syscall.SysProcAttr{Setsid: true},
			})
	case 1:
		return d.pseudoFork(nil)
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

// pseudoFork forks a daemon. It is supposed to be called from the daemon.
func (d *Daemon) pseudoFork(cmd *exec.Cmd) int {
	err := setArgs(
		cmd,
		d.Forked+1,
		d.BinPath,
		d.DbPath,
		d.SockPath,
		d.LogPathPrefix,
	).Start()

	if err != nil {
		return 2
	}
	return 0
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
