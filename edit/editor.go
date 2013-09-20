// package edit implements a full-feature line editor.
package edit

import (
	"os"
	"fmt"
	"unicode"
	"unicode/utf8"
	"./tty"
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
func Init(file *os.File) (*Editor, error) {
	fd := int(file.Fd())
	term, err := tty.NewTermiosFromFd(fd)
	if err != nil {
		return nil, fmt.Errorf("Can't get terminal attribute: %s", err)
	}

	editor := &Editor{
		savedTermios: term.Copy(),
		file: file,
		writer: newWriter(),
		reader: newReader(file),
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

func (ed *Editor) refresh(prompt, text, tip string) error {
	return ed.writer.refresh(prompt, text, tip, ed.file)
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine(prompt string) (lr LineRead) {
	line := ""
	tip := ""

	for {
		err := ed.refresh(prompt, line, tip)
		if err != nil {
			return LineRead{Err: err}
		}

		k, err := ed.reader.readKey()
		if err != nil {
			return LineRead{Err: err}
		}

		switch {
		case k == PlainKey('\n'):
			tip = ""
			err := ed.refresh(prompt, line, tip)
			if err != nil {
				return LineRead{Err: err}
			}
			fmt.Fprintln(ed.file)
			return LineRead{Line: line}
		case k == PlainKey(Backspace): // Backspace
			if l := len(line); l > 0 {
				_, w := utf8.DecodeLastRuneInString(line)
				line = line[:l-w]
			} else {
				ed.beep()
			}
		case k == CtrlKey('U'):
			line = ""
		case k == CtrlKey('D') && len(line) == 0:
			return LineRead{Eof: true}
		case k == CtrlKey('B'):
			fmt.Fprintf(ed.file, "\033[D")
		case k == CtrlKey('F'):
			fmt.Fprintf(ed.file, "\033[C")
		case !(k.Ctrl || k.Alt) && unicode.IsGraphic(k.rune):
			line += string(k.rune)
		default:
			tip = fmt.Sprintf("Unknown: %v", k)
		}
	}
}
