package main

import (
	"os"
	"fmt"
	"../parse"
	"../edit"
	"../eval"
	"../util"
)

func lackeol() {
	fmt.Println("\033[7m%\033[m")
}

func main() {
	tr, err := util.NewTimedReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	ev := eval.NewEvaluator(os.Environ())
	cmd_no := 0

	for {
		cmd_no++
		name := fmt.Sprintf("<tty %d>", cmd_no)

		ed, err := edit.Init(os.Stdin, tr, ev)
		if err != nil {
			panic(err)
		}

		prompt := util.Getwd() + "> "
		lr := ed.ReadLine(prompt)
		err = ed.Cleanup()
		if err != nil {
			panic(err)
		}

		if lr.Eof {
			lackeol()
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
