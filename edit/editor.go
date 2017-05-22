// Package edit implements a command line editor.
package edit

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[edit] ")

const (
	lackEOLRune = '\u23ce'
	lackEOL     = "\033[7m" + string(lackEOLRune) + "\033[m"
)

// Editor keeps the status of the line editor.
type Editor struct {
	file   *os.File
	writer *Writer
	reader *tty.Reader
	sigs   chan os.Signal
	store  *store.Store
	evaler *eval.Evaler
	cmdSeq int

	variables map[string]eval.Variable

	active       bool
	activeMutex  sync.Mutex
	historyMutex sync.RWMutex

	editorState
}

type editorState struct {
	// States used during ReadLine. Reset at the beginning of ReadLine.
	savedTermios *sys.Termios

	notificationMutex sync.Mutex

	notifications []string
	tips          []string

	line           string
	lexedLine      *string
	chunk          *parse.Chunk
	styling        *styling
	promptContent  []*styled
	rpromptContent []*styled
	dot            int

	mode Mode

	insert     insert
	command    command
	completion completion
	navigation navigation
	hist       hist
	histlist   *histlist
	bang       *bang
	location   *location

	// A cache of external commands, used in stylist.
	isExternal      map[string]bool
	parseErrorAtEnd bool

	// Used for builtins.
	lastKey    ui.Key
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

	ed := &Editor{
		file:   file,
		writer: newWriter(file),
		reader: tty.NewReader(file),
		sigs:   sigs,
		store:  st,
		evaler: ev,
		cmdSeq: seq,

		variables: makeVariables(),
	}
	ev.Editor = ed

	installModules(ev.Modules, ed)

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

func (ed *Editor) refresh(fullRefresh bool, addErrorsToTips bool) error {
	src := ed.line
	// Re-lex the line if needed
	if ed.lexedLine == nil || *ed.lexedLine != src {
		ed.lexedLine = &src
		n, err := parse.Parse("[interactive]", src)
		ed.chunk = n

		ed.parseErrorAtEnd = err != nil && atEnd(err, len(src))
		// If all parse errors are at the end, it is likely caused by incomplete
		// input. In that case, do not complain about parse errors.
		// TODO(xiaq): Find a more reliable way to determine incomplete input.
		// Ideally the parser should report it.
		if err != nil && addErrorsToTips && !ed.parseErrorAtEnd {
			ed.addTip("%s", err)
		}

		ed.styling = &styling{}
		highlight(n, ed)

		_, err = ed.evaler.Compile(n, "[interactive]", src)
		if err != nil && !atEnd(err, len(src)) {
			if addErrorsToTips {
				ed.addTip("%s", err)
			}
			// Highlight errors in the input buffer.
			// TODO(xiaq): There might be multiple tokens involved in the
			// compiler error; they should all be highlighted as erroneous.
			p := err.(*eval.CompilationError).Context.Begin
			badn := findLeafNode(n, p)
			ed.styling.add(badn.Begin(), badn.End(), styleForCompilerError.String())
		}
	}
	return ed.writer.refresh(&ed.editorState, fullRefresh)
}

func atEnd(e error, n int) bool {
	switch e := e.(type) {
	case *eval.CompilationError:
		return e.Context.Begin == n
	case *parse.Error:
		for _, entry := range e.Entries {
			if entry.Context.Begin != n {
				return false
			}
		}
		return true
	default:
		logger.Printf("atEnd called with error type %T", e)
		return false
	}
}

// insertAtDot inserts text at the dot and moves the dot after it.
func (ed *Editor) insertAtDot(text string) {
	ed.line = ed.line[:ed.dot] + text + ed.line[ed.dot:]
	ed.dot += len(text)
}

const flushInputDuringSetup = false

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

	if flushInputDuringSetup {
		err = sys.FlushInput(fd)
		if err != nil {
			return nil, fmt.Errorf("can't flush input: %s", err)
		}
	}

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
	/*
		Write a lackEOLRune if the cursor is not in the leftmost column. This is
		done as follows:

		1. Turn on autowrap;

		2. Write lackEOL along with enough padding, so that the total width is
		   equal to the width of the screen.

		   If the cursor was in the first column, we are still in the same line,
		   just off the line boundary. Otherwise, we are now in the next line.

		3. Rewind to the first column, write one space and rewind again. If the
		   cursor was in the first column to start with, we have just erased the
		   LackEOL character. Otherwise, we are now in the next line and this is
		   a no-op. The LackEOL character remains.
	*/
	fmt.Fprintf(ed.file, "\033[?7h%s%*s\r \r", lackEOL, width-util.Wcwidth(lackEOLRune), "")

	/*
		Turn off autowrap.

		The terminals sometimes has different opinions about how wide some
		characters are (notably emojis and some dingbats) with elvish. When that
		happens, elvish becomes wrong about where the cursor is when it writes
		its output, and the effect can be disastrous.

		If we turn off autowrap, the terminal won't insert any newlines behind
		the scene, so elvish is always right about which line the cursor is.
		With a bit more caution, this can restrict the consequence of the
		mismatch within one line.
	*/
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

	// Refresh the terminal for the last time in a clean-ish state.
	ed.mode = &ed.insert
	ed.tips = nil
	ed.dot = len(ed.line)
	if !ed.rpromptPersistent() {
		ed.rpromptContent = nil
	}
	addError(ed.refresh(false, false))
	ed.file.WriteString("\n")
	ed.writer.resetOldBuf()

	ed.reader.Quit()

	// Turn on autowrap.
	ed.file.WriteString("\033[?7h")
	// Turn off mouse tracking.
	//ed.file.WriteString("\033[?1000;1006l")

	// Disable bracketed paste.
	ed.file.WriteString("\033[?2004l")

	// Restore termios.
	err := ed.savedTermios.ApplyToFd(int(ed.file.Fd()))
	if err != nil {
		addError(fmt.Errorf("can't restore terminal attribute: %s", err))
	}

	// Save the line before resetting all of editorState.
	line := ed.line

	ed.editorState = editorState{}

	callHooks(ed.evaler, ed.afterReadLine(), eval.String(line))
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

	// Find external commands asynchronously, so that slow I/O won't block the
	// editor.
	isExternalCh := make(chan map[string]bool, 1)
	go getIsExternal(ed.evaler, isExternalCh)

	go ed.reader.Run()

	fullRefresh := false

	callHooks(ed.evaler, ed.beforeReadLine())

MainLoop:
	for {
		ed.promptContent = callPrompt(ed, ed.prompt())
		ed.rpromptContent = callPrompt(ed, ed.rprompt())

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
				continue MainLoop
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
		case unit := <-ed.reader.UnitChan():
			switch unit := unit.(type) {
			case tty.MouseEvent:
				ed.addTip("mouse: %+v", unit)
			case tty.CursorPosition:
				// Ignore CPR
			case tty.PasteSetting:
				if !unit {
					continue
				}
				var buf bytes.Buffer
				timer := time.NewTimer(tty.EscSequenceTimeout)
			paste:
				for {
					// XXX Should also select on other chans. However those chans
					// will be unified (again) into one later so we don't do
					// busywork here.
					select {
					case unit := <-ed.reader.UnitChan():
						switch unit := unit.(type) {
						case tty.Key:
							k := ui.Key(unit)
							if k.Mod != 0 {
								ed.Notify("function key within paste, aborting")
								break paste
							}
							buf.WriteRune(k.Rune)
							timer.Reset(tty.EscSequenceTimeout)
						case tty.PasteSetting:
							if !unit {
								break paste
							}
						default: // Ignore other things.
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
			case tty.RawRune:
				insertRaw(ed, rune(unit))
			case tty.Key:
				k := ui.Key(unit)
			lookupKey:
				keyBinding, ok := keyBindings[ed.mode.Mode()]
				if !ok {
					ed.addTip("No binding for current mode")
					continue
				}

				fn, bound := keyBinding[k]
				if !bound {
					// TODO(xiaq) don't assume Default always exists
					fn = keyBinding[ui.Default]
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
					err := ed.refresh(false, true)
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
}

// getIsExternal finds a set of all external commands and puts it on the result
// channel.
func getIsExternal(ev *eval.Evaler, result chan<- map[string]bool) {
	isExternal := make(map[string]bool)
	ev.EachExternal(func(name string) {
		isExternal[name] = true
	})
	result <- isExternal
}
