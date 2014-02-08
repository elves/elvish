// elvish is an experimental Unix shell. It tries to incorporate a powerful
// programming language with an extensible, friendly user interface.
package main

import (
	"fmt"
	"github.com/xiaq/elvish/edit"
	"github.com/xiaq/elvish/eval"
	"github.com/xiaq/elvish/parse"
	"github.com/xiaq/elvish/util"
	"os"
	"os/user"
)

func main() {
	tr, err := util.NewTimedReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	ev := eval.NewEvaluator(os.Environ())
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
	rprompt := username + "@" + hostname

	for {
		cmdNum++
		name := fmt.Sprintf("<tty %d>", cmdNum)

		ed, err := edit.Init(os.Stdin, tr, ev)
		if err != nil {
			panic(err)
		}

		prompt := util.Getwd() + "> "
		lr := ed.ReadLine(prompt, rprompt)
		err = ed.Cleanup()
		if err != nil {
			panic(err)
		}

		if lr.EOF {
			break
		} else if lr.Err != nil {
			panic(lr.Err)
		}

		n, pe := parse.Parse(name, lr.Line)
		if pe != nil {
			fmt.Print(pe.(*util.ContextualError).Pprint())
			continue
		}

		ee := ev.Eval(name, lr.Line, n)
		if ee != nil {
			fmt.Println(ee.(*util.ContextualError).Pprint())
			continue
		}
	}
}
