// Package edit implements a command line editor.
package edit

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var Logger = util.GetLogger("[edit] ")

const (
	lackEOLRune = '\u23ce'
	lackEOL     = "\033[7m" + string(lackEOLRune) + "\033[m"
)

// Editor keeps the status of the line editor.
type Editor struct {
	file   *os.File
	writer *writer
	reader *Reader
	sigs   chan os.Signal
	store  *store.Store
	evaler *eval.Evaler
	cmdSeq int

	ps1           eval.Variable
	rps1          eval.Variable
	completers    map[string]ArgCompleter
	abbreviations map[string]string

	rpromptPersistent bool
	beforeReadLine    eval.Variable
	afterReadLine     eval.Variable

	historyMutex sync.RWMutex

	active      bool
	activeMutex sync.Mutex

	editorState
}

type editorState struct {
	// States used during ReadLine. Reset at the beginning of ReadLine.
	savedTermios *sys.Termios

	notificationMutex sync.Mutex

	notifications []string
	tips          []string

	tokens  []Token
	prompt  []*styled
	rprompt []*styled
	line    string
	dot     int

	mode Mode

	insert     insert
	command    command
	completion completion
	navigation navigation
	hist       hist
	histlist   *histlist
	bang       *bang
	location   *location

	// A cache of external commands, used in stylist and completer of command
	// names.
	isExternal      map[string]bool
	parseErrorAtEnd bool

	// Used for builtins.
	lastKey    Key
	nextAction action
}

// NewEditor creates an Editor.
func NewEditor(file *os.File, sigs chan os.Signal, ev *eval.Evaler, st *store.Store) *Editor {
	seq := -1
	if st != nil {
		var err error
		seq, err = st.NextCmdSeq()
		if err != nil {
			// TODO(xiaq): Also report the error
			seq = -1
		}
	}

	prompt, rprompt := defaultPrompts()

	ed := &Editor{
		file:   file,
		writer: newWriter(file),
		reader: NewReader(file),
		sigs:   sigs,
		store:  st,
		evaler: ev,
		cmdSeq: seq,
		ps1:    eval.NewPtrVariableWithValidator(prompt, MustBeFn),
		rps1:   eval.NewPtrVariableWithValidator(rprompt, MustBeFn),

		abbreviations: make(map[string]string),
		beforeReadLine: eval.NewPtrVariableWithValidator(
			eval.NewList(), eval.IsListOfFnValue),
		afterReadLine: eval.NewPtrVariableWithValidator(
			eval.NewList(), eval.IsListOfFnValue),
	}
	ev.Editor = ed
	ev.Modules["le"] = makeModule(ed)
	return ed
}

func (ed *Editor) Active() bool {
	return ed.active
}

func (ed *Editor) ActiveMutex() *sync.Mutex {
	return &ed.activeMutex
}

func (ed *Editor) flash() {
	// TODO implement fish-like flash effect
}

func (ed *Editor) addTip(format string, args ...interface{}) {
	ed.tips = append(ed.tips, fmt.Sprintf(format, args...))
}

func (ed *Editor) Notify(format string, args ...interface{}) {
	ed.notificationMutex.Lock()
	defer ed.notificationMutex.Unlock()
	ed.notifications = append(ed.notifications, fmt.Sprintf(format, args...))
}

func (ed *Editor) refresh(fullRefresh bool, tips bool) error {
	// Re-lex the line, unless we are in modeCompletion
	src := ed.line
	if ed.mode.Mode() != modeCompletion {
		n, err := parse.Parse("[interactive]", src)
		ed.parseErrorAtEnd = err != nil && atEnd(err, len(src))
		if err != nil {
			// If all the errors happen at the end, it is liekly complaining
			// about missing texts that will eventually be inserted. Don't show
			// such errors.
			// XXX We may need a more reliable criteria.
			if tips && !ed.parseErrorAtEnd {
				ed.addTip("parser error: %s", err)
			}
		}
		if n == nil {
			ed.tokens = []Token{parserError(src, 0, len(src))}
		} else {
			ed.tokens = tokenize(src, n)
			_, err := ed.evaler.Compile(n, "[interactive]", src)
			if err != nil {
				if tips && !atEnd(err, len(src)) {
					ed.addTip("compiler error: %s", err)
				}
				if err, ok := err.(*util.PosError); ok {
					p := err.Begin
					for i, token := range ed.tokens {
						if token.Node.Begin() <= p && p < token.Node.End() {
							ed.tokens[i].MoreStyle = joinStyles(ed.tokens[i].MoreStyle, styleForCompilerError)
							break
						}
					}
				}
			}
		}
		stylist := &Stylist{ed.tokens, ed, nil}
		stylist.do(n)
	}
	return ed.writer.refresh(&ed.editorState, fullRefresh)
}

func atEnd(e error, n int) bool {
	switch e := e.(type) {
	case *util.PosError:
		return e.Begin == n
	case *util.Errors:
		for _, child := range e.Errors {
			if !atEnd(child, n) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// insertAtDot inserts text at the dot and moves the dot after it.
func (ed *Editor) insertAtDot(text string) {
	ed.line = ed.line[:ed.dot] + text + ed.line[ed.dot:]
	ed.dot += len(text)
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

	/*
		err = sys.FlushInput(fd)
		if err != nil {
			return nil, fmt.Errorf("can't flush input: %s", err)
		}
	*/

	return savedTermios, nil
}

// startReadLine prepares the terminal for the editor.
func (ed *Editor) startReadLine() error {
	ed.activeMutex.Lock()
	defer ed.activeMutex.Unlock()
	ed.active = true

	savedTermios, err := setupTerminal(ed.file)
	if err != nil {
		return err
	}
	ed.savedTermios = savedTermios

	_, width := sys.GetWinsize(int(ed.file.Fd()))
	// Turn on autowrap, write lackEOL along with enough padding to fill the
	// whole screen. If the cursor was in the first column, we end up in the
	// same line (just off the line boundary); otherwise we are now in the next
	// line. We now rewind to the first column and erase anything there. The
	// final effect is that a lackEOL gets written if and only if the cursor
	// was not in the first column.
	fmt.Fprintf(ed.file, "\033[?7h%s%*s\r \r", lackEOL, width-util.Wcwidth(lackEOLRune), "")

	// Turn off autowrap. The edito has its own wrapping mechanism. Doing
	// wrapping manually means that when the actual width of some characters
	// are greater than what our wcwidth implementation tells us, characters at
	// the end of that line gets hidden -- compared to pushed to the next line,
	// which is more disastrous.
	ed.file.WriteString("\033[?7l")
	// Turn on SGR-style mouse tracking.
	//ed.file.WriteString("\033[?1000;1006h")

	// Enable bracketed paste.
	ed.file.WriteString("\033[?2004h")

	return nil
}

// finishReadLine puts the terminal in a state suitable for other programs to
// use.
func (ed *Editor) finishReadLine(addError func(error)) {
	ed.activeMutex.Lock()
	defer ed.activeMutex.Unlock()
	ed.active = false

	ed.mode = &ed.insert
	ed.tips = nil
	ed.dot = len(ed.line)
	if !ed.rpromptPersistent {
		ed.rprompt = nil
	}
	addError(ed.refresh(false, false))
	ed.file.WriteString("\n")

	// ed.reader.Stop()
	ed.reader.Quit()

	// Turn on autowrap.
	ed.file.WriteString("\033[?7h")
	// Turn off mouse tracking.
	//ed.file.WriteString("\033[?1000;1006l")

	// Disable bracketed paste.
	ed.file.WriteString("\033[?2004l")

	// restore termios
	err := ed.savedTermios.ApplyToFd(int(ed.file.Fd()))

	if err != nil {
		addError(fmt.Errorf("can't restore terminal attribute: %s", err))
	}
	ed.savedTermios = nil
	line := ed.line
	ed.editorState = editorState{}

	afterReadLine := ed.afterReadLine.Get().(eval.ListLike)
	if afterReadLine.Len() > 0 {
		opfunc := func(ec *eval.EvalCtx) {
			afterReadLine.Iterate(func(v eval.Value) bool {
				fn := v.(eval.FnValue)
				ex := ec.PCall(fn, []eval.Value{eval.String(line)}, eval.NoOpts)
				// TODO Pretty print.
				if ex != nil {
					fmt.Fprintln(os.Stderr, "function error: %s", ex.Error())
				}
				return true
			})
		}
		ed.evaler.Eval(eval.Op{opfunc, -1, -1}, "[after-readline]", "no source")
	}
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine() (line string, err error) {
	e := ed.startReadLine()
	if e != nil {
		return "", e
	}
	defer ed.finishReadLine(func(e error) {
		if e != nil {
			err = util.CatError(err, e)
		}
	})

	ed.mode = &ed.insert

	isExternalCh := make(chan map[string]bool, 1)
	go getIsExternal(ed.evaler, isExternalCh)

	ed.writer.resetOldBuf()
	go ed.reader.Run()

	fullRefresh := false

	beforeReadLines := ed.beforeReadLine.Get().(eval.ListLike)
	beforeReadLines.Iterate(func(f eval.Value) bool {
		ed.CallFn(f.(eval.FnValue))
		return true
	})

MainLoop:
	for {
		ed.prompt = callFnForPrompt(ed, ed.ps1.Get().(eval.Fn))
		ed.rprompt = callFnForPrompt(ed, ed.rps1.Get().(eval.Fn))

		err := ed.refresh(fullRefresh, true)
		fullRefresh = false
		if err != nil {
			return "", err
		}

		ed.tips = nil

		select {
		case m := <-isExternalCh:
			ed.isExternal = m
		case sig := <-ed.sigs:
			// TODO(xiaq): Maybe support customizable handling of signals
			switch sig {
			case syscall.SIGINT:
				// Start over
				ed.editorState = editorState{
					savedTermios: ed.savedTermios,
					isExternal:   ed.isExternal,
				}
				ed.mode = &ed.insert
				goto MainLoop
			case syscall.SIGWINCH:
				fullRefresh = true
				continue MainLoop
			case syscall.SIGCHLD:
				// ignore
			default:
				ed.addTip("ignored signal %s", sig)
			}
		case err := <-ed.reader.ErrorChan():
			ed.Notify("reader error: %s", err.Error())
		case mouse := <-ed.reader.MouseChan():
			ed.addTip("mouse: %+v", mouse)
		case <-ed.reader.CPRChan():
			// Ignore CPR
		case b := <-ed.reader.PasteChan():
			if !b {
				continue
			}
			var buf bytes.Buffer
			timer := time.NewTimer(EscSequenceTimeout)
		paste:
			for {
				// XXX Should also select on other chans. However those chans
				// will be unified (agina) into one later so we don't do
				// busywork here.
				select {
				case k := <-ed.reader.KeyChan():
					if k.Mod != 0 {
						ed.Notify("function key within paste")
						break paste
					}
					buf.WriteRune(k.Rune)
					timer.Reset(EscSequenceTimeout)
				case b := <-ed.reader.PasteChan():
					if !b {
						break paste
					}
				case <-timer.C:
					ed.Notify("bracketed paste timeout")
					break paste
				}
			}
			topaste := buf.String()
			if ed.insert.quotePaste {
				topaste = parse.Quote(topaste)
			}
			ed.insertAtDot(topaste)
		case k := <-ed.reader.KeyChan():
		lookupKey:
			keyBinding, ok := keyBindings[ed.mode.Mode()]
			if !ok {
				ed.addTip("No binding for current mode")
				continue
			}

			fn, bound := keyBinding[k]
			if !bound {
				fn = keyBinding[Default]
			}

			ed.insert.insertedLiteral = false
			ed.lastKey = k
			ed.CallFn(fn)
			if ed.insert.insertedLiteral {
				ed.insert.literalInserts++
			} else {
				ed.insert.literalInserts = 0
			}
			act := ed.nextAction
			ed.nextAction = action{}

			switch act.typ {
			case noAction:
				continue
			case reprocessKey:
				err = ed.refresh(false, true)
				if err != nil {
					return "", err
				}
				goto lookupKey
			case exitReadLine:
				if act.returnErr == nil && act.returnLine != "" {
					ed.appendHistory(act.returnLine)
				}
				return act.returnLine, act.returnErr
			}
		}
	}
}

// getIsExternal finds a set of all external commands and puts it on the result
// channel.
func getIsExternal(ev *eval.Evaler, result chan<- map[string]bool) {
	names := make(chan string, 32)
	go func() {
		ev.AllExecutables(names)
		close(names)
	}()
	isExternal := make(map[string]bool)
	for name := range names {
		isExternal[name] = true
	}
	result <- isExternal
}
