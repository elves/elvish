// package edit implements a full-feature line editor.
package edit

import (
	"os"
	"fmt"
	"unicode"
	"unicode/utf8"
	"./tty"
	"../eval"
	"../util"
)

var lackeol = "\033[7m\u23ce\033[m\n"

// Editor keeps the status of the line editor.
type Editor struct {
	savedTermios *tty.Termios
	file *os.File
	writer *writer
	reader *reader
	// Fields below are used when during ReadLine.
	prompt, line, tip string
	completions []string
	currentCompletion int
	dot int
}

// LineRead is the result of ReadLine. Exactly one member is non-zero, making
// it effectively a tagged union.
type LineRead struct {
	Line string
	Eof bool
	Err error
}

// Init initializes an Editor on the terminal referenced by fd.
func Init(file *os.File, tr *util.TimedReader, ev *eval.Evaluator) (*Editor, error) {
	fd := int(file.Fd())
	term, err := tty.NewTermiosFromFd(fd)
	if err != nil {
		return nil, fmt.Errorf("Can't get terminal attribute: %s", err)
	}

	editor := &Editor{
		savedTermios: term.Copy(),
		file: file,
		writer: newWriter(file, ev),
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

	err = tty.FlushInput(fd)
	if err != nil {
		return nil, err
	}

	file.WriteString("\033[6n")
	// XXX Possible race condition: user input sneaked in between WriteString
	// and readCPR
	x, _, err := editor.reader.readCPR()
	if err != nil {
		return nil, err
	}

	if x != 1 {
		file.WriteString(lackeol)
	}

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

func (ed *Editor) pushTip(more string) {
	if len(ed.tip) == 0 {
		ed.tip = more
	} else {
		ed.tip = ed.tip + "; " + more
	}
}

func (ed *Editor) refresh() error {
	return ed.writer.refresh(ed.prompt, ed.line, ed.tip, ed.completions, ed.currentCompletion, ed.dot)
}

// TODO Allow modifiable keybindings.
var keyBindings = map[Key]string {
	Key{'U', Ctrl}: "kill-line-b",
	Key{'K', Ctrl}: "kill-line-f",
	Key{Backspace, 0}: "kill-rune-b",
	Key{Left, 0}: "move-dot-b",
	Key{Right, 0}: "move-dot-f",
	Key{Tab, 0}: "complete",
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine(prompt string) (lr LineRead) {
	ed.prompt = prompt
	ed.line = ""
	ed.tip = ""
	ed.completions = nil
	ed.dot = 0

	for {
		err := ed.refresh()
		if err != nil {
			return LineRead{Err: err}
		}

		ed.tip = ""

		k, err := ed.reader.readKey()
		if err != nil {
			ed.pushTip(err.Error())
			continue
		}

		if name, bound := keyBindings[k]; bound {
			leBuiltins[name](ed)
			continue
		}

		switch k {
		// XXX Keybindings that affect the flow of ReadLine can't yet be
		// implemented as functions.
		case Key{Enter, 0}:
			err := ed.refresh()
			if err != nil {
				return LineRead{Err: err}
			}
			fmt.Fprintln(ed.file)
			return LineRead{Line: ed.line}
		case Key{'D', Ctrl}:
			if len(ed.line) == 0 {
				return LineRead{Eof: true}
			}
			fallthrough
		default:
			if k.Mod == 0 && k.rune > 0 && unicode.IsGraphic(k.rune) {
				ed.line = ed.line[:ed.dot] + string(k.rune) + ed.line[ed.dot:]
				ed.dot += utf8.RuneLen(k.rune)
			} else {
				ed.pushTip(fmt.Sprintf("Unbound: %s", k))
			}
		}
	}
}
