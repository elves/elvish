// Package edit implements a command line editor.
package edit

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sys"
)

const lackEOL = "\033[7m\u23ce\033[m\n"

type bufferMode int

const (
	modeInsert bufferMode = iota
	modeCommand
	modeCompletion
	modeNavigation
	modeHistory
)

type editorState struct {
	// States used during ReadLine. Reset at the beginning of ReadLine.
	savedTermios          *sys.Termios
	lastKey               Key
	tokens                []Token
	prompt, rprompt, line string
	dot                   int
	tips                  []string
	mode                  bufferMode
	completion            *completion
	completionLines       int
	navigation            *navigation
	history               history
	isExternal            map[string]bool
}

type history struct {
	current int
	prefix  string
	line    string
}

// Editor keeps the status of the line editor.
type Editor struct {
	file      *os.File
	writer    *writer
	reader    *Reader
	sigs      <-chan os.Signal
	histories []string
	store     *store.Store
	evaler    *eval.Evaler
	cmdSeq    int
	editorState
}

// LineRead is the result of ReadLine. Exactly one member is non-zero, making
// it effectively a tagged union.
type LineRead struct {
	Line string
	EOF  bool
	Err  error
}

func (h *history) jump(i int, line string) {
	h.current = i
	h.line = line
}

func (ed *Editor) appendHistory(line string) {
	ed.histories = append(ed.histories, line)
	if ed.store != nil {
		ed.store.AddCmd(line)
		// TODO(xiaq): Report possible error
	}
}

func lastHistory(histories []string, upto int, prefix string) (int, string) {
	for i := upto - 1; i >= 0; i-- {
		if strings.HasPrefix(histories[i], prefix) {
			return i, histories[i]
		}
	}
	return -1, ""
}

func firstHistory(histories []string, from int, prefix string) (int, string) {
	for i := from; i < len(histories); i++ {
		if strings.HasPrefix(histories[i], prefix) {
			return i, histories[i]
		}
	}
	return -1, ""
}

func (ed *Editor) prevHistory() bool {
	if ed.history.current > 0 {
		// Session history
		i, line := lastHistory(ed.histories, ed.history.current, ed.history.prefix)
		if i >= 0 {
			ed.history.jump(i, line)
			return true
		}
	}

	if ed.store != nil {
		// Persistent history
		upto := ed.cmdSeq + min(0, ed.history.current)
		i, line, err := ed.store.LastCmd(upto, ed.history.prefix)
		if err == nil {
			ed.history.jump(i-ed.cmdSeq, line)
			return true
		}
	}
	// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
	return false
}

func (ed *Editor) nextHistory() bool {
	if ed.store != nil {
		// Persistent history
		if ed.history.current < -1 {
			from := ed.cmdSeq + ed.history.current + 1
			i, line, err := ed.store.FirstCmd(from, ed.history.prefix)
			if err == nil {
				ed.history.jump(i-ed.cmdSeq, line)
				return true
			}
			// TODO(xiaq): Errors other than ErrNoMatchingCmd should be reported
		}
	}

	from := max(0, ed.history.current+1)
	i, line := firstHistory(ed.histories, from, ed.history.prefix)
	if i >= 0 {
		ed.history.jump(i, line)
		return true
	}
	return false
}

// NewEditor creates an Editor.
func NewEditor(file *os.File, sigs <-chan os.Signal, ev *eval.Evaler, st *store.Store) *Editor {
	seq := -1
	if st != nil {
		var err error
		seq, err = st.NextCmdSeq()
		if err != nil {
			// TODO(xiaq): Also report the error
			seq = -1
		}
	}

	return &Editor{
		file:   file,
		writer: newWriter(file),
		reader: NewReader(file),
		sigs:   sigs,
		store:  st,
		evaler: ev,
		cmdSeq: seq,
	}
}

func (ed *Editor) flash() {
	// TODO implement fish-like flash effect
}

func (ed *Editor) pushTip(more string) {
	ed.tips = append(ed.tips, more)
}

func (ed *Editor) refresh() error {
	ed.reader.Stop()
	defer ed.reader.Continue()
	// Re-lex the line, unless we are in modeCompletion
	if ed.mode != modeCompletion {
		// XXX Ignore error
		ed.tokens, _ = tokenize(ed.line)
		for i, t := range ed.tokens {
			for _, colorist := range colorists {
				ed.tokens[i].MoreStyle += colorist(t.Node, ed)
			}
		}
	}
	return ed.writer.refresh(&ed.editorState)
}

// TODO Allow modifiable keybindings.
var keyBindings = map[bufferMode]map[Key]string{
	modeCommand: map[Key]string{
		Key{'i', 0}:    "start-insert",
		Key{'h', 0}:    "move-dot-left",
		Key{'l', 0}:    "move-dot-right",
		Key{'D', 0}:    "kill-line-right",
		DefaultBinding: "default-command",
	},
	modeInsert: map[Key]string{
		Key{'[', Ctrl}:    "start-command",
		Key{'U', Ctrl}:    "kill-line-left",
		Key{'K', Ctrl}:    "kill-line-right",
		Key{'W', Ctrl}:    "kill-word-left",
		Key{Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		Key{'H', Ctrl}:  "kill-rune-left",
		Key{Delete, 0}:  "kill-rune-right",
		Key{Left, 0}:    "move-dot-left",
		Key{Right, 0}:   "move-dot-right",
		Key{Up, 0}:      "move-dot-up",
		Key{Down, 0}:    "move-dot-down",
		Key{Enter, Alt}: "insert-key",
		Key{Enter, 0}:   "return-line",
		Key{'D', Ctrl}:  "return-eof",
		Key{Tab, 0}:     "start-completion",
		Key{PageUp, 0}:  "start-history",
		Key{'N', Ctrl}:  "start-navigation",
		DefaultBinding:  "default-insert",
	},
	modeCompletion: map[Key]string{
		Key{'[', Ctrl}: "cancel-completion",
		Key{Up, 0}:     "select-cand-up",
		Key{Down, 0}:   "select-cand-down",
		Key{Left, 0}:   "select-cand-left",
		Key{Right, 0}:  "select-cand-right",
		// Key{Tab, 0}:    "cycle-cand-right",
		DefaultBinding: "default-completion",
	},
	modeNavigation: map[Key]string{
		Key{Up, 0}:     "select-nav-up",
		Key{Down, 0}:   "select-nav-down",
		Key{Left, 0}:   "ascend-nav",
		Key{Right, 0}:  "descend-nav",
		DefaultBinding: "default-navigation",
	},
	modeHistory: map[Key]string{
		Key{'[', Ctrl}:   "start-insert",
		Key{PageUp, 0}:   "select-history-prev",
		Key{PageDown, 0}: "select-history-next",
		DefaultBinding:   "default-history",
	},
}

func init() {
	for _, kb := range keyBindings {
		for _, name := range kb {
			if leBuiltins[name] == nil {
				panic("bad keyBindings table: no editor builtin named " + name)
			}
		}
	}
}

// acceptCompletion accepts currently selected completion candidate.
func (ed *Editor) acceptCompletion() {
	c := ed.completion
	if 0 <= c.current && c.current < len(c.candidates) {
		accepted := c.candidates[c.current].source.text
		// Insert the accepted completion text at ed.dot and move ed.dot after
		// the newly inserted text
		ed.line = ed.line[:ed.dot] + accepted + ed.line[ed.dot:]
		ed.dot += len(accepted)
	}
	ed.completion = nil
	ed.mode = modeInsert
}

// acceptHistory accepts currently history.
func (ed *Editor) acceptHistory() {
	ed.line = ed.history.line
	ed.dot = len(ed.line)
}

func setupTerminal(file *os.File) (*sys.Termios, error) {
	fd := int(file.Fd())
	term, err := sys.NewTermiosFromFd(fd)
	if err != nil {
		return nil, fmt.Errorf("can't get terminal attribute: %s", err)
	}

	savedTermios := term.Copy()

	term.SetICanon(false)
	term.SetEcho(false)
	term.SetVMin(1)
	term.SetVTime(0)

	err = term.ApplyToFd(fd)
	if err != nil {
		return nil, fmt.Errorf("can't set up terminal attribute: %s", err)
	}

	// Set autowrap off
	file.WriteString("\033[?7l")

	err = sys.FlushInput(fd)
	if err != nil {
		return nil, fmt.Errorf("can't flush input: %s", err)
	}

	return savedTermios, nil
}

func cleanupTerminal(file *os.File, savedTermios *sys.Termios) error {
	// Set autowrap on
	file.WriteString("\033[?7h")
	fd := int(file.Fd())
	return savedTermios.ApplyToFd(fd)
}

// startsReadLine prepares the terminal for the editor.
func (ed *Editor) startReadLine() error {
	savedTermios, err := setupTerminal(ed.file)
	if err != nil {
		return err
	}
	ed.savedTermios = savedTermios

	// Query cursor location
	ed.file.WriteString("\033[6n")

	ed.reader.Continue()
	ones := ed.reader.Chan()

	cpr := invalidPos
FindCPR:
	for {
		select {
		case or := <-ones:
			if or.CPR != invalidPos {
				cpr = or.CPR
				break FindCPR
			} else {
				// Just discard
			}
		case <-time.After(cprTimeout):
			break FindCPR
		}
	}

	if cpr == invalidPos {
		// Unable to get CPR, just rewind to column 1
		ed.file.WriteString("\r")
	} else if cpr.col != 1 {
		// BUG(xiaq) startReadline assumes that column number starts from 0
		ed.file.WriteString(lackEOL)
	}

	return nil
}

// finishReadLine puts the terminal in a state suitable for other programs to
// use.
func (ed *Editor) finishReadLine(lr *LineRead) {
	if lr.EOF == false && lr.Err == nil && lr.Line != "" {
		ed.appendHistory(lr.Line)
	}

	ed.mode = modeInsert
	ed.tips = nil
	ed.completion = nil
	ed.navigation = nil
	ed.dot = len(ed.line)
	// TODO Perhaps make it optional to NOT clear the rprompt
	ed.rprompt = ""
	ed.refresh() // XXX(xiaq): Ignore possible error
	ed.file.WriteString("\n")

	ed.reader.Stop()

	err := cleanupTerminal(ed.file, ed.savedTermios)

	if err != nil {
		// BUG(xiaq): Error in Editor.finishReadLine may override earlier error
		*lr = LineRead{Err: fmt.Errorf("can't restore terminal attribute: %s", err)}
	}
	ed.savedTermios = nil
}

// ReadLine reads a line interactively.
// TODO(xiaq): ReadLine currently handles SIGINT and SIGWINCH and swallows all
// other signals.
func (ed *Editor) ReadLine(prompt, rprompt func() string) (lr LineRead) {
	ed.editorState = editorState{}
	go ed.updateIsExternal()

	ed.writer.oldBuf.cells = nil
	ones := ed.reader.Chan()

	err := ed.startReadLine()
	if err != nil {
		return LineRead{Err: err}
	}
	defer ed.finishReadLine(&lr)

MainLoop:
	for {
		ed.prompt = prompt()
		ed.rprompt = rprompt()

		err := ed.refresh()
		if err != nil {
			return LineRead{Err: err}
		}

		ed.tips = nil

		select {
		case sig := <-ed.sigs:
			// TODO(xiaq): Maybe support customizable handling of signals
			switch sig {
			case syscall.SIGINT:
				// Start over
				ed.editorState = editorState{savedTermios: ed.savedTermios}
				goto MainLoop
			case syscall.SIGWINCH:
				continue MainLoop
			}
		case or := <-ones:
			// Alert about error
			err := or.Err
			if err != nil {
				ed.pushTip(err.Error())
				continue
			}

			// Ignore bogus CPR
			if or.CPR != invalidPos {
				continue
			}

			k := or.Key
		lookupKey:
			keyBinding, ok := keyBindings[ed.mode]
			if !ok {
				ed.pushTip("No binding for current mode")
				continue
			}

			name, bound := keyBinding[k]
			if !bound {
				name = keyBinding[DefaultBinding]
			}
			ed.lastKey = k
			ret := leBuiltins[name](ed)
			if ret == nil {
				continue
			}
			switch ret.action {
			case noAction:
				continue
			case reprocessKey:
				err = ed.refresh()
				if err != nil {
					return LineRead{Err: err}
				}
				goto lookupKey
			case exitReadLine:
				return ret.readLineReturn
			}
		}
	}
}
