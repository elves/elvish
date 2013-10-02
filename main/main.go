package main

import (
	"os"
	"fmt"
	"syscall"
	"../parse"
	"../edit"
	"../eval"
	"../async"
)

func lackeol() {
	fmt.Println("\033[7m%\033[m")
}

func main() {
	fmt.Printf("My pid is %d\n", syscall.Getpid())

	cmd_no := 0

	tr, err := async.NewTimedReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	for {
		cmd_no++
		name := fmt.Sprintf("<interactive code %d>", cmd_no)

		ed, err := edit.Init(os.Stdin, tr)
		if err != nil {
			panic(err)
		}

		lr := ed.ReadLine("das> ")
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

		tree, err := parse.Parse(name, lr.Line, false)
		if err != nil {
			fmt.Println("Parser error:", err)
			continue
		}

		updates, err := eval.ExecPipeline(tree.Root.(*parse.ListNode))
		if err != nil {
			fmt.Println(err)
			continue
		}

		for i, update := range updates {
			// TODO Should check update.Msg
			for up := range update {
				switch up.Msg {
				case "0", "":
				default:
					fmt.Printf("Command #%d update: %s\n", i, up.Msg)
				}
			}
		}
	}
}
