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
	dot := 0

	for {
		err := ed.writer.refresh(prompt, line, tip, dot)
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
			err := ed.writer.refresh(prompt, line, tip, dot)
			if err != nil {
				return LineRead{Err: err}
			}
			fmt.Fprintln(ed.file)
			return LineRead{Line: line}
		case Key{Backspace, 0}:
			if dot > 0 {
				_, w := utf8.DecodeLastRuneInString(line[:dot])
				line = line[:dot-w] + line[dot:]
				dot -= w
			} else {
				ed.beep()
			}
		case Key{'U', Ctrl}:
			line = line[dot:]
			dot = 0
		case Key{Left, 0}:
			_, w := utf8.DecodeLastRuneInString(line[:dot])
			dot -= w
		case Key{Right, 0}:
			_, w := utf8.DecodeRuneInString(line[dot:])
			dot += w
		case Key{'D', Ctrl}:
			if len(line) == 0 {
				return LineRead{Eof: true}
			}
			fallthrough
		default:
			if k.Mod == 0 && unicode.IsGraphic(k.rune) {
				line = line[:dot] + string(k.rune) + line[dot:]
				dot += utf8.RuneLen(k.rune)
			} else {
				tip = pushTip(tip, fmt.Sprintf("Unbound: %s", k))
			}
		}
	}
}
