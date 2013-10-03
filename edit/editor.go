// package edit implements a full-feature line editor.
package edit

import (
	"os"
	"fmt"
	"unicode"
	"unicode/utf8"
	"./tty"
	"../async"
)

// Editor keeps the status of the line editor.
type Editor struct {
	savedTermios *tty.Termios
	file *os.File
	writer *writer
	reader *reader
}

// LineRead is the result of ReadLine. Exactly one member is non-zero, making
// it effectively a tagged union.
type LineRead struct {
	Line string
	Eof bool
	Err error
}

// Init initializes an Editor on the terminal referenced by fd.
func Init(file *os.File, tr *async.TimedReader) (*Editor, error) {
	fd := int(file.Fd())
	term, err := tty.NewTermiosFromFd(fd)
	if err != nil {
		return nil, fmt.Errorf("Can't get terminal attribute: %s", err)
	}

	editor := &Editor{
		savedTermios: term.Copy(),
		file: file,
		writer: newWriter(file),
		reader: newReader(tr),
	}

	term.SetIcanon(false)
	term.SetEcho(false)
	term.SetMin(1)
	term.SetTime(0)

	err = term.ApplyToFd(fd)
	if err != nil {
		return nil, fmt.Errorf("Can't set up terminal attribute: %s", err)
	}

	fmt.Fprint(editor.file, "\033[?7l")
	return editor, nil
}

// Cleanup restores the terminal referenced by fd so that other commands
// that use the terminal can be executed.
func (ed *Editor) Cleanup() error {
	fmt.Fprint(ed.file, "\033[?7h")

	fd := int(ed.file.Fd())
	err := ed.savedTermios.ApplyToFd(fd)
	if err != nil {
		return fmt.Errorf("Can't restore terminal attribute of stdin: %s", err)
	}
	ed.savedTermios = nil
	return nil
}

func (ed *Editor) beep() {
}

func (ed *Editor) refresh(prompt, text, tip string, point int) error {
	return ed.writer.refresh(prompt, text, tip, point)
}

func pushTip(tip, more string) string {
	if len(tip) == 0 {
		return more
	}
	return tip + "; " + more
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine(prompt string) (lr LineRead) {
	line := ""
	tip := ""
	point := 0

	for {
		err := ed.refresh(prompt, line, tip, point)
		if err != nil {
			return LineRead{Err: err}
		}

		tip = ""

		k, err := ed.reader.readKey()
		if err != nil {
			tip = pushTip(tip, err.Error())
		}

		switch k {
		case Key{Enter, 0}:
			tip = ""
			err := ed.refresh(prompt, line, tip, point)
			if err != nil {
				return LineRead{Err: err}
			}
			fmt.Fprintln(ed.file)
			return LineRead{Line: line}
		case Key{Backspace, 0}:
			if l := len(line); l > 0 {
				_, w := utf8.DecodeLastRuneInString(line)
				line = line[:l-w]
				point--
			} else {
				ed.beep()
			}
		case Key{'U', Ctrl}:
			line = ""
			point = 0
		case Key{'D', Ctrl}:
			if len(line) == 0 {
				return LineRead{Eof: true}
			}
			fallthrough
		default:
			if k.Mod == 0 && unicode.IsGraphic(k.rune) {
				line += string(k.rune)
				point++
			} else {
				tip = pushTip(tip, fmt.Sprintf("Unbound: %s", k))
			}
		}
	}
}
