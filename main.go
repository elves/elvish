// elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/errutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sysutil"
)

const (
	sigchSize     = 32
	outChanSize   = 32
	outChanLeader = "▶ "
)

func newEvalerAndStore() (*eval.Evaler, *store.Store) {
	dataDir, err := store.EnsureDataDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: cannot create data dir ~/.elvish")
	}

	var st *store.Store
	if err == nil {
		st, err = store.NewStore(dataDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: cannot connect to store:", err)
		}
	}

	return eval.NewEvaler(st, dataDir), st
}

func printError(err error) {
	if err != nil {
		if ce, ok := err.(*errutil.ContextualError); ok {
			fmt.Fprint(os.Stderr, ce.Pprint())
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

// TODO(xiaq): Currently only the editor deals with signals.
func interact() {
	ev, st := newEvalerAndStore()
	datadir, err := store.EnsureDataDir()
	printError(err)
	if err == nil {
		// XXX
		vs, err := ev.Source(datadir + "/rc.elv")
		if err != nil && !os.IsNotExist(err) {
			printError(err)
		}
		eval.PrintExituses(vs)
	}

	cmdNum := 0

	username := "???"
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "???"
	}
	rpromptStr := username + "@" + hostname

	sigch := make(chan os.Signal, sigchSize)

	ed := edit.NewEditor(os.Stdin, sigch, ev, st)

	for {
		cmdNum++
		name := fmt.Sprintf("<tty %d>", cmdNum)

		prompt := func() string {
			return sysutil.Getwd() + "> "
		}
		rprompt := func() string {
			return rpromptStr
		}

		signal.Notify(sigch)
		lr := ed.ReadLine(prompt, rprompt)
		signal.Stop(sigch)

		if lr.EOF {
			break
		} else if lr.Err != nil {
			fmt.Println("Editor error:", lr.Err)
			fmt.Println("My pid is", os.Getpid())
		}

		n, err := parse.Parse(name, lr.Line)
		printError(err)

		if err == nil {
			vs, err := ev.Eval(name, lr.Line, n)
			printError(err)
			eval.PrintExituses(vs)
		}
	}
}

func script(fname string) {
	ev, _ := newEvalerAndStore()
	vs, err := ev.Source(fname)
	printError(err)
	eval.PrintExituses(vs)
	if err != nil || eval.HasFailure(vs) {
		os.Exit(1)
	}
}

var usage = `Usage:
    elvish
    elvish <script>
`

func main() {
	switch len(os.Args) {
	case 1:
		interact()
	case 2:
		script(os.Args[1])
	default:
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}
}
