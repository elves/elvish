// package edit implements a full-feature line editor.
package edit

import (
	"os"
	"fmt"
	"bufio"
	"unicode"
	"unicode/utf8"
	"./tty"
)

// Editor keeps the status of the line editor.
type Editor struct {
	savedTermios *tty.Termios
	file *os.File
	writer *writer
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

func (ed *Editor) tip(s string) {
	fmt.Fprintf(ed.file, "\n%s\033[A", s)
}

func (ed *Editor) tipf(format string, a ...interface{}) {
	ed.tip(fmt.Sprintf(format, a...))
}

func (ed *Editor) clearTip() {
	fmt.Fprintf(ed.file, "\n\033[K\033[A")
}

func (ed *Editor) refresh(prompt, text string) error {
	return ed.writer.refresh(prompt, text, ed.file)
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine(prompt string) (lr LineRead) {
	stdin := bufio.NewReaderSize(ed.file, 0)
	line := ""

	for {
		err := ed.refresh(prompt, line)
		if err != nil {
			return LineRead{Err: err}
		}

		r, _, err := stdin.ReadRune()
		if err != nil {
			return LineRead{Err: err}
		}

		switch {
		case r == '\n':
			ed.clearTip()
			fmt.Fprintln(ed.file)
			return LineRead{Line: line}
		case r == 0x7f: // Backspace
			if l := len(line); l > 0 {
				_, w := utf8.DecodeLastRuneInString(line)
				line = line[:l-w]
			} else {
				ed.beep()
			}
		case r == 0x15: // ^U
			line = ""
		case r == 0x4 && len(line) == 0: // ^D
			return LineRead{Eof: true}
		case r == 0x2: // ^B
			fmt.Fprintf(ed.file, "\033[D")
		case r == 0x6: // ^F
			fmt.Fprintf(ed.file, "\033[C")
		case unicode.IsGraphic(r):
			line += string(r)
		default:
			ed.tipf("Non-graphic: %#x", r)
		}
	}
}
