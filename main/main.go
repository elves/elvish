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

	rr := async.NewRuneReader(os.Stdin, 0)

	for {
		cmd_no++
		name := fmt.Sprintf("<interactive code %d>", cmd_no)

		ed, err := edit.Init(os.Stdin, rr)
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

		pids, err := eval.ExecPipeline(tree.Root.(*parse.ListNode))
		if err != nil {
			fmt.Println(err)
			if pids == nil {
				continue
			}
		}

		var ws syscall.WaitStatus
		var ru syscall.Rusage

		// TODO Should check ws
		for _, pid := range pids {
			syscall.Wait4(pid, &ws, 0, &ru)
		}
	}
}
