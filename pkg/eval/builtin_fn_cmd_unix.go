//go:build unix

package eval

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/sys/eunix"
)

// ErrNotInSameProcessGroup is thrown when the process IDs passed to fg are not
// in the same process group.
var ErrNotInSameProcessGroup = errors.New("not in the same process group")

// Reference to syscall.Exec. Can be overridden in tests.
var syscallExec = syscall.Exec

func execFn(fm *Frame, args ...any) error {
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

	fm.Evaler.PreExit()
	decSHLVL()

	return syscallExec(argstrings[0], argstrings, os.Environ())
}

// Decrements $E:SHLVL. Called from execFn to ensure that $E:SHLVL remains the
// same in the new command.
func decSHLVL() {
	i, err := strconv.Atoi(os.Getenv(env.SHLVL))
	if err != nil {
		return
	}
	os.Setenv(env.SHLVL, strconv.Itoa(i-1))
}

func fg(pids ...int) error {
	if len(pids) == 0 {
		return errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: -1, Actual: len(pids)}
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
			return ErrNotInSameProcessGroup
		}
	}

	err := eunix.Tcsetpgrp(0, thepgid)
	if err != nil {
		return err
	}

	errors := make([]Exception, len(pids))

	for i, pid := range pids {
		err := syscall.Kill(pid, syscall.SIGCONT)
		if err != nil {
			errors[i] = &exception{err, nil}
		}
	}

	for i, pid := range pids {
		if errors[i] != nil {
			continue
		}
		var ws syscall.WaitStatus
		_, err = syscall.Wait4(pid, &ws, syscall.WUNTRACED, nil)
		if err != nil {
			errors[i] = &exception{err, nil}
		} else {
			// TODO find command name
			errors[i] = &exception{NewExternalCmdExit(
				"[pid "+strconv.Itoa(pid)+"]", ws, pid), nil}
		}
	}

	return MakePipelineError(errors)
}
