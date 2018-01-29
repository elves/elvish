// +build !windows,!plan9

package eval

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/sys"
)

func execFn(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	var argstrings []string
	if len(args) == 0 {
		argstrings = []string{"elvish"}
	} else {
		argstrings = make([]string, len(args))
		for i, a := range args {
			argstrings[i] = types.ToString(a)
		}
	}

	var err error
	argstrings[0], err = exec.LookPath(argstrings[0])
	maybeThrow(err)

	preExit(ec)

	err = syscall.Exec(argstrings[0], argstrings, os.Environ())
	maybeThrow(err)
}

func fg(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var pids []int
	ScanArgsVariadic(args, &pids)
	TakeNoOpt(opts)

	if len(pids) == 0 {
		throw(ErrArgs)
	}
	var thepgid int
	for i, pid := range pids {
		pgid, err := syscall.Getpgid(pid)
		maybeThrow(err)
		if i == 0 {
			thepgid = pgid
		} else if pgid != thepgid {
			throw(ErrNotInSameGroup)
		}
	}

	err := sys.Tcsetpgrp(0, thepgid)
	maybeThrow(err)

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

	maybeThrow(ComposeExceptionsFromPipeline(errors))
}
