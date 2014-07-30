// elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"unicode/utf8"

	"github.com/xiaq/elvish/edit"
	"github.com/xiaq/elvish/eval"
	"github.com/xiaq/elvish/parse"
	"github.com/xiaq/elvish/util"
)

const (
	sigchSize = 32
)

// TODO(xiaq): Currently only the editor deals with signals.
func interact() {
	ev := eval.NewEvaluator()
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
		if pe != nil {
			fmt.Print(pe.(*util.ContextualError).Pprint())
			continue
		}

		ee := ev.Eval(name, lr.Line, n)
		if ee != nil {
			if ce, ok := ee.(*util.ContextualError); ok {
				fmt.Print(ce.Pprint())
			} else {
				fmt.Println(ee)
			}
			continue
		}
	}
}

func script(name string) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !utf8.Valid(bytes) {
		fmt.Fprintf(os.Stderr, "source %v is not valid UTF-8\n", name)
		os.Exit(1)
	}
	src := string(bytes)

	ev := eval.NewEvaluator()

	n, pe := parse.Parse(name, src)
	if pe != nil {
		fmt.Print(pe.(*util.ContextualError).Pprint())
		os.Exit(1)
	}

	ee := ev.Eval(name, src, n)
	if ee != nil {
		fmt.Print(ee.(*util.ContextualError).Pprint())
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
