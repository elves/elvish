// +build !windows,!plan9

package eval

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/sys"
)

func execFn(fm *Frame, args ...interface{}) error {
	var argstrings []string
	if len(args) == 0 {
		argstrings = []string{"elvish"}
	} else {
		argstrings = make([]string, len(args))
		for i, a := range args {
			argstrings[i] = vals.ToString(a)
		}
	}

	var err error
	argstrings[0], err = exec.LookPath(argstrings[0])
	if err != nil {
		return err
	}

	preExit(fm)

	return syscall.Exec(argstrings[0], argstrings, os.Environ())
}

func fg(pids ...int) error {
	if len(pids) == 0 {
		return ErrArgs
	}
	var thepgid int
	for i, pid := range pids {
		pgid, err := syscall.Getpgid(pid)
		if err != nil {
			return err
		}
		if i == 0 {
			thepgid = pgid
		} else if pgid != thepgid {
			return ErrNotInSameGroup
		}
	}

	err := sys.Tcsetpgrp(0, thepgid)
	if err != nil {
		return err
	}

	errors := make([]*Exception, len(pids))

	for i, pid := range pids {
		err := syscall.Kill(pid, syscall.SIGCONT)
		if err != nil {
			errors[i] = &Exception{err, nil}
		}
	}

	for i, pid := range pids {
		if errors[i] != nil {
			continue
		}
		var ws syscall.WaitStatus
		_, err = syscall.Wait4(pid, &ws, syscall.WUNTRACED, nil)
		if err != nil {
			errors[i] = &Exception{err, nil}
		} else {
			// TODO find command name
			errors[i] = &Exception{NewExternalCmdExit(
				"[pid "+strconv.Itoa(pid)+"]", ws, pid), nil}
		}
	}

	return ComposeExceptionsFromPipeline(errors)
}
