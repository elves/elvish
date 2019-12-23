package main

import (
	"log"
	"os"

	"github.com/elves/elvish/pkg/sys"
	"golang.org/x/sys/windows"
)

func main() {
	restore := setup(os.Stdin, os.Stdout)
	defer restore()

	log.Println("ready")
	console, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		log.Fatalf("GetStdHandle(STD_INPUT_HANDLE): %v", err)
	}
	for {
		event, err := sys.ReadInputEvent(console)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("input: %#v", event)
	}
}

const (
	wantedInMode = windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT
	wantedOutMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
)

func setup(in, out *os.File) func() {
	hIn := windows.Handle(in.Fd())
	hOut := windows.Handle(out.Fd())

	var oldInMode, oldOutMode uint32
	err := windows.GetConsoleMode(hIn, &oldInMode)
	if err != nil {
		log.Fatal(err)
	}
	err = windows.GetConsoleMode(hOut, &oldOutMode)
	if err != nil {
		log.Fatal(err)
	}

	err = windows.SetConsoleMode(hIn, wantedInMode)
	if err != nil {
		log.Fatal(err)
	}
	err = windows.SetConsoleMode(hOut, wantedOutMode)
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		err := windows.SetConsoleMode(hIn, oldInMode)
		if err != nil {
			log.Fatal(err)
		}
		err = windows.SetConsoleMode(hOut, oldOutMode)
		if err != nil {
			log.Fatal(err)
		}
	}
}
