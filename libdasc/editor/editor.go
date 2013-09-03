// package editor implements a full-feature line editor.
package editor

import (
	"os"
	"fmt"
	"bufio"
	"unicode"
	"unicode/utf8"
	"./tty"
)

type LineRead struct {
	Line string
	Eof bool
	Err error
}

var savedTermios *tty.Termios

func Init() error {
	var err error
	savedTermios, err = tty.NewTermiosFromFd(0)
	if err != nil {
		return fmt.Errorf("Can't get terminal attribute of stdin: %s", err)
	}
	term := savedTermios.Copy()

	term.SetIcanon(false)
	term.SetEcho(false)
	term.SetMin(1)
	term.SetTime(0)

	err = term.ApplyToFd(0)
	if err != nil {
		return fmt.Errorf("Can't set up terminal attribute of stdin: %s", err)
	}

	fmt.Print("\033[?7l")
	return nil
}

func Cleanup() error {
	fmt.Print("\033[?7h")

	err := savedTermios.ApplyToFd(0)
	if err != nil {
		return fmt.Errorf("Can't restore terminal attribute of stdin: %s", err)
	}
	savedTermios = nil
	return nil
}

func beep() {
}

func tip(s string) {
	fmt.Printf("\n%s\033[A", s)
}

func tipf(format string, a ...interface{}) {
	tip(fmt.Sprintf(format, a...))
}

func clearTip() {
	fmt.Printf("\n\033[K\033[A")
}

func refresh(prompt, text string) (newlines int, err error) {
	w := newWriter()
	defer func() {
		newlines = w.line
	}()
	for _, r := range prompt {
		err = w.write(r)
		if err != nil {
			return
		}
	}
	var indent int
	if w.col * 2 < w.width {
		indent = w.col
	}
	for _, r := range text {
		err = w.write(r)
		if err != nil {
			return
		}
		if w.col == 0 {
			for i := 0; i < indent; i++ {
				err = w.write(' ')
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func ReadLine(prompt string) (lr LineRead) {
	stdin := bufio.NewReaderSize(os.Stdin, 0)
	line := ""

	newlines := 0

	for {
		if newlines > 0 {
			fmt.Printf("\033[%dA", newlines)
		}
		fmt.Printf("\r\033[J")

		newlines, _ = refresh(prompt, line)

		r, _, err := stdin.ReadRune()
		if err != nil {
			return LineRead{Err: err}
		}

		switch {
		case r == '\n':
			clearTip()
			fmt.Println()
			return LineRead{Line: line}
		case r == 0x7f: // Backspace
			if l := len(line); l > 0 {
				_, w := utf8.DecodeLastRuneInString(line)
				line = line[:l-w]
			} else {
				beep()
			}
		case r == 0x15: // ^U
			line = ""
		case r == 0x4 && len(line) == 0: // ^D
			return LineRead{Eof: true}
		case r == 0x2: // ^B
			fmt.Printf("\033[D")
		case r == 0x6: // ^F
			fmt.Printf("\033[C")
		case unicode.IsGraphic(r):
			line += string(r)
		default:
			tipf("Non-graphic: %#x", r)
		}
	}

	panic("unreachable")
}
