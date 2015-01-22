// elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"

	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	_ "github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

const (
	sigchSize     = 32
	outChanSize   = 32
	outChanLeader = "▶ "
)

func newEvaluator() *eval.Evaluator {
	ch := make(chan eval.Value, outChanSize)
	go func() {
		for v := range ch {
			fmt.Printf("%s%s\n", outChanLeader, v.Repr())
		}
	}()

	ev := eval.NewEvaluator()
	ev.SetChanOut(ch)
	return ev
}

func printError(err error) {
	if err != nil {
		if ce, ok := err.(*util.ContextualError); ok {
			fmt.Fprint(os.Stderr, ce.Pprint())
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

// TODO(xiaq): Currently only the editor deals with signals.
func interact() {
	ev := newEvaluator()

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
	signal.Notify(sigch)

	ed := edit.NewEditor(os.Stdin, ev, sigch)

	for {
		cmdNum++
		name := fmt.Sprintf("<tty %d>", cmdNum)

		prompt := func() string {
			return util.Getwd() + "> "
		}
		rprompt := func() string {
			return rpromptStr
		}

		lr := ed.ReadLine(prompt, rprompt)

		if lr.EOF {
			break
		} else if lr.Err != nil {
			fmt.Println("Editor error:", lr.Err)
			fmt.Println("My pid is", os.Getpid())
		}

		n, pe := parse.Parse(name, lr.Line)
		printError(pe)

		ee := ev.Eval(name, lr.Line, n)
		printError(ee)
	}
}

func script(fname string) {
	ev := newEvaluator()
	err := ev.Source(fname)
	printError(err)
	if err != nil {
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
