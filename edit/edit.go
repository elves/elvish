// Package edit implements a command line editor.
package edit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/edit/highlight"
	"github.com/elves/elvish/edit/history"
	"github.com/elves/elvish/edit/prompt"
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[edit] ")

// Editor keeps the status of the line editor.
type Editor struct {
	in     *os.File
	out    *os.File
	writer tty.Writer
	reader tty.Reader
	sigs   chan os.Signal
	daemon *daemon.Client
	evaler *eval.Evaler

	variables map[string]vartypes.Variable
	bindings  map[string]vartypes.Variable

	active      bool
	activeMutex sync.Mutex

	historyFuser *history.Fuser
	historyMutex sync.RWMutex

	// notifyPort is a write-only port that turns data written to it into editor
	// notifications.
	notifyPort *eval.Port
	// notifyRead is the read end of notifyPort.File.
	notifyRead *os.File

	editorState
}

type editorState struct {
	// States used during ReadLine. Reset at the beginning of ReadLine.
	restoreTerminal func() error

	notificationMutex sync.Mutex

	notifications []string
	tips          []string

	buffer string
	dot    int

	chunk           *parse.Chunk
	styling         *highlight.Styling
	parseErrorAtEnd bool

	promptContent  []*ui.Styled
	rpromptContent []*ui.Styled

	mode Mode

	insert     insert
	command    command
	completion completion
	navigation navigation

	// A cache of external commands, used in stylist.
	isExternal map[string]bool

	// Used for builtins.
	lastKey    ui.Key
	nextAction action
}

// NewEditor creates an Editor. When the instance is no longer used, its Close
// method should be called.
func NewEditor(in *os.File, out *os.File, sigs chan os.Signal, ev *eval.Evaler) *Editor {
	daemon := ev.DaemonClient

	ed := &Editor{
		in:     in,
		out:    out,
		writer: tty.NewWriter(out),
		reader: tty.NewReader(in),
		sigs:   sigs,
		daemon: daemon,
		evaler: ev,

		bindings:  makeBindings(),
		variables: makeVariables(),
	}

	notifyChan := make(chan types.Value)
	notifyRead, notifyWrite, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	ed.notifyPort = &eval.Port{File: notifyWrite, Chan: notifyChan}
	ed.notifyRead = notifyRead
	// Forward reads from notifyRead to notification.
	go func() {
		reader := bufio.NewReader(notifyRead)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			ed.Notify("[bytes out] %s", line[:len(line)-1])
		}
		if err != io.EOF {
			logger.Println("notifyRead error:", err)
		}
	}()
	// Forward reads from notifyChan to notification.
	go func() {
		for v := range notifyChan {
			ed.Notify("[value out] %s", types.Repr(v, types.NoPretty))
		}
	}()

	if daemon != nil {
		f, err := history.NewFuser(daemon)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to initialize command history. Disabled.")
		} else {
			ed.historyFuser = f
		}
	}
	ev.Editor = ed

	installModules(ev.Builtin, ed)
	err = ev.SourceText(eval.NewScriptSource("[editor]", "[editor]", "use binding; binding:install"))
	if err != nil {
		fmt.Fprintln(out, "Failed to load default binding:", err)
	}

	return ed
}

// Close releases resources used by the editor.
func (ed *Editor) Close() {
	ed.reader.Close()
	close(ed.notifyPort.Chan)
	ed.notifyPort.File.Close()
	ed.notifyRead.Close()
}

// Active returns the activeness of the Editor.
func (ed *Editor) Active() bool {
	return ed.active
}

// ActiveMutex returns a mutex that must be used when changing the activeness of
// the Editor.
func (ed *Editor) ActiveMutex() *sync.Mutex {
	return &ed.activeMutex
}

func (ed *Editor) Evaler() *eval.Evaler {
	return ed.evaler
}

func (ed *Editor) Variable(name string) vartypes.Variable {
	return ed.variables[name]
}

func (ed *Editor) flash() {
	// TODO implement fish-like flash effect
}

func (ed *Editor) addTip(format string, args ...interface{}) {
	ed.tips = append(ed.tips, fmt.Sprintf(format, args...))
}

// Notify adds one notification entry. It is concurrency-safe.
func (ed *Editor) Notify(format string, args ...interface{}) {
	ed.notificationMutex.Lock()
	defer ed.notificationMutex.Unlock()
	ed.notifications = append(ed.notifications, fmt.Sprintf(format, args...))
}

func (ed *Editor) refresh(fullRefresh bool, addErrorsToTips bool) error {
	src := ed.buffer
	// Parse the current line
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

	ed.styling = &highlight.Styling{}
	doHighlight(n, ed)

	_, err = ed.evaler.Compile(n, eval.NewInteractiveSource(src))
	if err != nil && !atEnd(err, len(src)) {
		if addErrorsToTips {
			ed.addTip("%s", err)
		}
		// Highlight errors in the input buffer.
		ctx := err.(*eval.CompilationError).Context
		ed.styling.Add(ctx.Begin, ctx.End, styleForCompilerError.String())
	}

	// Render onto a buffer.
	height, width := sys.GetWinsize(ed.out)
	height = min(height, ed.maxHeight())
	er := &editorRenderer{&ed.editorState, height, nil}
	buf := ui.Render(er, width)
	return ed.writer.CommitBuffer(er.bufNoti, buf, fullRefresh)
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
	ed.buffer = ed.buffer[:ed.dot] + text + ed.buffer[ed.dot:]
	ed.dot += len(text)
}

// startReadLine prepares the terminal for the editor.
func (ed *Editor) startReadLine() error {
	ed.activeMutex.Lock()
	defer ed.activeMutex.Unlock()
	ed.active = true

	restoreTerminal, err := tty.Setup(ed.in, ed.out)
	if err != nil {
		if restoreTerminal != nil {
			restoreTerminal()
		}
		return err
	}
	ed.restoreTerminal = restoreTerminal

	return nil
}

// finishReadLine puts the terminal in a state suitable for other programs to
// use.
func (ed *Editor) finishReadLine() error {
	ed.activeMutex.Lock()
	defer ed.activeMutex.Unlock()
	ed.active = false

	// Refresh the terminal for the last time in a clean-ish state.
	ed.mode = &ed.insert
	ed.tips = nil
	ed.dot = len(ed.buffer)
	if !prompt.RpromptPersistent(ed) {
		ed.rpromptContent = nil
	}
	errRefresh := ed.refresh(false, false)
	ed.out.WriteString("\n")
	ed.writer.ResetCurrentBuffer()

	ed.reader.Stop()

	// Restore termios.
	errRestore := ed.restoreTerminal()

	// Save the line before resetting all of editorState.
	line := ed.buffer
	ed.editorState = editorState{}

	callHooks(ed.evaler, ed.afterReadLine(), line)

	return util.Errors(errRefresh, errRestore)
}

// ReadLine reads a line interactively.
func (ed *Editor) ReadLine() (string, error) {
	err := ed.startReadLine()
	if err != nil {
		return "", err
	}
	defer func() {
		err := ed.finishReadLine()
		if err != nil {
			fmt.Fprintln(ed.out, "error:", err)
		}
	}()

	ed.mode = &ed.insert

	// Find external commands asynchronously, so that slow I/O won't block the
	// editor.
	isExternalCh := make(chan map[string]bool, 1)
	go getIsExternal(ed.evaler, isExternalCh)

	ed.reader.Start()

	fullRefresh := false

	callHooks(ed.evaler, ed.beforeReadLine())

	promptUpdater := prompt.NewUpdater(prompt.Prompt)
	rpromptUpdater := prompt.NewUpdater(prompt.Rprompt)

MainLoop:
	for {
		promptCh := promptUpdater.Update(ed)
		rpromptCh := rpromptUpdater.Update(ed)
		promptTimeout := prompt.MakeMaxWaitChan(ed)
		rpromptTimeout := prompt.MakeMaxWaitChan(ed)

		select {
		case ed.promptContent = <-promptCh:
			logger.Println("prompt fetched")
		case <-promptTimeout:
			logger.Println("stale prompt")
			ed.promptContent = promptUpdater.Staled
		}
		select {
		case ed.rpromptContent = <-rpromptCh:
			logger.Println("rprompt fetched")
		case <-rpromptTimeout:
			logger.Println("stale rprompt")
			ed.rpromptContent = rpromptUpdater.Staled
		}

	refresh:
		err := ed.refresh(fullRefresh, true)
		fullRefresh = false
		if err != nil {
			return "", err
		}

		ed.tips = nil

		select {
		case ed.promptContent = <-promptCh:
			logger.Println("prompt fetched late")
			goto refresh
		case ed.rpromptContent = <-rpromptCh:
			logger.Println("rprompt fetched late")
			goto refresh
		case m := <-isExternalCh:
			ed.isExternal = m
		case sig := <-ed.sigs:
			// TODO(xiaq): Maybe support customizable handling of signals
			switch sig {
			case syscall.SIGHUP:
				return "", io.EOF
			case syscall.SIGINT:
				// Start over
				ed.editorState = editorState{
					restoreTerminal: ed.restoreTerminal,
					isExternal:      ed.isExternal,
				}
				ed.mode = &ed.insert
				continue MainLoop
			case sys.SIGWINCH:
				fullRefresh = true
				continue MainLoop
			default:
				ed.addTip("ignored signal %s", sig)
			}
		case event := <-ed.reader.EventChan():
			switch event := event.(type) {
			case tty.NonfatalErrorEvent:
				ed.Notify("error when reading terminal: %v", event.Err)
			case tty.FatalErrorEvent:
				ed.Notify("fatal error when reading terminal: %v", event.Err)
				return "", event.Err
			case tty.MouseEvent:
				ed.addTip("mouse: %+v", event)
			case tty.CursorPosition:
				// Ignore CPR
			case tty.PasteSetting:
				if !event {
					continue
				}
				var buf bytes.Buffer
				timer := time.NewTimer(tty.DefaultSeqTimeout)
			paste:
				for {
					// XXX Should also select on other chans. However those chans
					// will be unified (again) into one later so we don't do
					// busywork here.
					select {
					case event := <-ed.reader.EventChan():
						switch event := event.(type) {
						case tty.KeyEvent:
							k := ui.Key(event)
							if k.Mod != 0 {
								ed.Notify("function key within paste, aborting")
								break paste
							}
							buf.WriteRune(k.Rune)
							timer.Reset(tty.DefaultSeqTimeout)
						case tty.PasteSetting:
							if !event {
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
				insertRaw(ed, rune(event))
			case tty.KeyEvent:
				k := ui.Key(event)
			lookupKey:
				fn := ed.mode.Binding(ed.bindings, k)
				if fn == nil {
					ed.addTip("Unbound and no default binding: %s", k)
					continue MainLoop
				}

				ed.insert.insertedLiteral = false
				ed.lastKey = k
				ed.CallFn(fn)
				if ed.insert.insertedLiteral {
					ed.insert.literalInserts++
				} else {
					ed.insert.literalInserts = 0
				}

				switch ed.popAction() {
				case reprocessKey:
					err := ed.refresh(false, true)
					if err != nil {
						return "", err
					}
					goto lookupKey
				case commitLine:
					ed.appendHistory(ed.buffer)
					return ed.buffer, nil
				case commitEOF:
					return "", io.EOF
				}
			}
		}
	}
}

// getIsExternal finds a set of all external commands and puts it on the result
// channel.
func getIsExternal(ev *eval.Evaler, result chan<- map[string]bool) {
	isExternal := make(map[string]bool)
	eval.EachExternal(func(name string) {
		isExternal[name] = true
	})
	result <- isExternal
}
