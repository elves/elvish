package main

import (
	"fmt"
	"github.com/xiaq/das/edit"
	"github.com/xiaq/das/eval"
	"github.com/xiaq/das/parse"
	"github.com/xiaq/das/util"
	"os"
	"os/user"
)

func main() {
	tr, err := util.NewTimedReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	ev := eval.NewEvaluator(os.Environ())
	cmd_no := 0

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
		cmd_no++
		name := fmt.Sprintf("<tty %d>", cmd_no)

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

		if lr.Eof {
			break
		} else if lr.Err != nil {
			panic(lr.Err)
		}

		p := parse.NewParser(name)
		tree, pe := p.Parse(lr.Line, false)
		if pe != nil {
			fmt.Print(pe.(*util.ContextualError).Pprint())
			continue
		}

		ee := ev.Eval(name, lr.Line, tree.Root)
		if ee != nil {
			fmt.Println(ee.(*util.ContextualError).Pprint())
			continue
		}
	}
}
