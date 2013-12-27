// package edit implements a full-feature line editor.
package edit

import (
	"os"
	"fmt"
	"unicode"
	"unicode/utf8"
	"./tty"
	"../eval"
	"../parse"
	"../util"
)

var Lackeol = "\033[7m\u23ce\033[m\n"

// Editor keeps the status of the line editor.
type Editor struct {
	savedTermios *tty.Termios
	file *os.File
	writer *writer
	reader *reader
	ev *eval.Evaluator
	// Fields below are used during ReadLine.
	tokens []parse.Item
	prompt, line, tip string
	completion *completion
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
// The Editor is reinitialized every time the control of the terminal is
// transferred back to the line editor.
func Init(file *os.File, tr *util.TimedReader, ev *eval.Evaluator) (*Editor, error) {
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
		ev: ev,
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
		file.WriteString(Lackeol)
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
	return ed.writer.refresh(ed.prompt, ed.tokens, ed.tip, ed.completion, ed.dot)
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

var completionKeyBindings = map[Key]string {
	Key{'[', Ctrl}: "cancel-completion",
	Key{Up, 0}: "select-cand-b",
	Key{Down, 0}: "select-cand-f",
	Key{Tab, 0}: "cycle-cand-f",
}

// Accpet currently selected completion candidate.
func (ed *Editor) acceptCompletion() {
	c := ed.completion
	if 0 <= c.current && c.current < len(c.candidates) {
		accepted := c.candidates[c.current].text
		ed.line = ed.line[:c.start] + accepted + ed.line[c.end:]
		ed.dot += len(accepted) - (c.end - c.start)
	}
	ed.completion = nil
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine(prompt string) (lr LineRead) {
	ed.prompt = prompt
	ed.line = ""
	ed.tip = ""
	ed.completion = nil
	ed.dot = 0

	for {
		// Re-lex the line, unless we are in completion mode
		if ed.completion == nil {
			ed.tokens = nil
			hl := Highlight("<interactive code>", ed.line, ed.ev)
			for token := range hl {
				ed.tokens = append(ed.tokens, token)
			}
		}

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

		if ed.completion != nil {
			if name, bound := completionKeyBindings[k]; bound {
				leBuiltins[name](ed)
				continue
			} else {
				// Implicitly accept completion and fall back to normal keybinding
				ed.acceptCompletion()
			}
		}

		if name, bound := keyBindings[k]; bound {
			leBuiltins[name](ed)
			continue
		}

		switch k {
		// XXX Keybindings that affect the flow of ReadLine can't yet be
		// implemented as functions.
		case Key{Enter, 0}:
			ed.tip = ""
			ed.completion = nil
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
