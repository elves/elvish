package main

import (
	"log"
	"os"
	"strings"
	"unicode"

	"golang.org/x/sys/windows"

	"src.elv.sh/pkg/sys/ewindows"
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
		var buf [1]ewindows.InputRecord
		nr, err := ewindows.ReadConsoleInput(console, buf[:])
		if nr == 0 {
			log.Fatal("no event read")
		}
		if err != nil {
			log.Fatal(err)
		}
		event := buf[0].GetEvent()
		switch event := event.(type) {
		case *ewindows.KeyEvent:
			typ := "up"
			if event.BKeyDown != 0 {
				typ = "down"
			}
			r := rune(event.UChar[0]) + rune(event.UChar[1])<<8
			rs := "(" + string(r) + ")"
			if unicode.IsControl(r) {
				rs = "   "
			}
			var mods []string
			testMod := func(mask uint32, name string) {
				if event.DwControlKeyState&mask != 0 {
					mods = append(mods, name)
				}
			}
			testMod(0x1, "right alt")
			testMod(0x2, "left alt")
			testMod(0x4, "right ctrl")
			testMod(0x8, "left ctrl")
			testMod(0x10, "shift")
			// testMod(0x20, "numslock")
			testMod(0x40, "scrolllock")
			testMod(0x80, "capslock")
			testMod(0x100, "enhanced")

			log.Printf("%4s, key code = %02x, char = %04x %s, mods = %s\n",
				typ, event.WVirtualKeyCode, r, rs, strings.Join(mods, ", "))
		}
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
