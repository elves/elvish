// Package edcore implements the core of the Elvish command editor.
package edcore

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
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/highlight"
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
)

var logger = util.GetLogger("[edit] ")

// editor implements the line editor.
type editor struct {
	in     *os.File
	out    *os.File
	writer tty.Writer
	reader tty.Reader
	sigs   <-chan os.Signal
	daemon *daemon.Client
	evaler *eval.Evaler

	active      bool
	activeMutex sync.Mutex

	// notifyPort is a write-only port that turns data written to it into editor
	// notifications.
	notifyPort *eval.Port
	// notifyRead is the read end of notifyPort.File.
	notifyRead *os.File

	// Configurations. Each of the following fields have an initializer defined
	// using atEditorInit.
	editorHooks
	abbr      hashmap.Map
	maxHeight float64

	prompt, rprompt   eddefs.Prompt
	RpromptPersistent bool

	// Modes.
	insert     *insert
	command    *command
	navigation *navigation
	listing    *listingMode

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

	promptContent, rpromptContent []*ui.Styled

	mode eddefs.Mode

	// A cache of external commands, used in stylist.
	isExternal map[string]bool

	// Used for builtins.
	lastKey    ui.Key
	nextAction eddefs.Action
}

// NewEditor creates an Editor. When the instance is no longer used, its Close
// method should be called.
func NewEditor(in *os.File, out *os.File, sigs <-chan os.Signal, ev *eval.Evaler) eddefs.Editor {
	daemon := ev.DaemonClient

	ed := &editor{
		in:     in,
		out:    out,
		writer: tty.NewWriter(out),
		reader: tty.NewReader(in),
		sigs:   sigs,
		daemon: daemon,
		evaler: ev,
	}

	notifyChan := make(chan interface{})
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
			ed.Notify("[value out] %s", vals.Repr(v, vals.NoPretty))
		}
	}()

	ev.Editor = ed

	ns := makeNs(ed)
	for _, f := range editorInitFuncs {
		f(ed, ns)
	}
	ev.Builtin.AddNs("edit", ns)

	err = ev.EvalSource(eval.NewScriptSource("[editor]", "[editor]", "use binding; binding:install"))
	if err != nil {
		fmt.Fprintln(out, "Failed to load default binding:", err)
	}

	return ed
}

func (ed *editor) Close() {
	ed.reader.Close()
	close(ed.notifyPort.Chan)
	ed.notifyPort.File.Close()
	ed.notifyRead.Close()
	ed.prompt.Close()
	ed.rprompt.Close()
}

func (ed *editor) Evaler() *eval.Evaler {
	return ed.evaler
}

func (ed *editor) Daemon() *daemon.Client {
	return ed.daemon
}

func (ed *editor) Buffer() (string, int) {
	return ed.buffer, ed.dot
}

func (ed *editor) SetBuffer(buffer string, dot int) {
	ed.buffer, ed.dot = buffer, dot
}

func (ed *editor) ParsedBuffer() *parse.Chunk {
	return ed.chunk
}

func (ed *editor) SetMode(m eddefs.Mode) {
	if ed.mode != nil {
		ed.mode.Teardown()
	}
	ed.mode = m
}

func (ed *editor) SetModeInsert() {
	ed.SetMode(ed.insert)
}

func (ed *editor) SetModeListing(b eddefs.BindingMap, p eddefs.ListingProvider) {
	ed.listing.setup(b, p)
	ed.SetMode(ed.listing)
}

func (ed *editor) RefreshListing() {
	if l, ok := ed.mode.(*listingMode); ok {
		l.refresh()
	}
}

func (ed *editor) flash() {
	// TODO implement fish-like flash effect
}

// AddTip adds a message to the tip area.
func (ed *editor) AddTip(format string, args ...interface{}) {
	ed.tips = append(ed.tips, fmt.Sprintf(format, args...))
}

// Notify writes out a message in a way that does not interrupt the editor
// display. When the editor is not active, it simply writes the message to the
// terminal. When the editor is active, it appends the message to the
// notification queue, which will be written out during the update cycle. It can
// be safely used concurrently.
func (ed *editor) Notify(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ed.activeMutex.Lock()
	defer ed.activeMutex.Unlock()
	// If the editor is not active, simply write out the message.
	if !ed.active {
		ed.out.WriteString(msg + "\n")
		return
	}
	ed.notificationMutex.Lock()
	defer ed.notificationMutex.Unlock()
	ed.notifications = append(ed.notifications, msg)
}

func (ed *editor) LastKey() ui.Key {
	return ed.lastKey
}

func (ed *editor) refresh(fullRefresh bool, addErrorsToTips bool) error {
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
		ed.AddTip("%s", err)
	}

	ed.styling = &highlight.Styling{}
	doHighlight(n, ed)

	_, err = ed.evaler.Compile(n, eval.NewInteractiveSource(src))
	if err != nil && !atEnd(err, len(src)) {
		if addErrorsToTips {
			ed.AddTip("%s", err)
		}
		// Highlight errors in the input buffer.
		ctx := err.(*eval.CompilationError).Context
		ed.styling.Add(ctx.Begin, ctx.End, styleForCompilerError.String())
	}

	// Render onto a buffer.
	height, width := sys.GetWinsize(ed.out)
	height = min(height, maxHeightToInt(ed.maxHeight))
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

// InsertAtDot inserts text at the dot and moves the dot after it.
func (ed *editor) InsertAtDot(text string) {
	ed.buffer = ed.buffer[:ed.dot] + text + ed.buffer[ed.dot:]
	ed.dot += len(text)
}

func (ed *editor) SetPrompt(prompt eddefs.Prompt) {
	ed.prompt = prompt
}

func (ed *editor) SetRPrompt(rprompt eddefs.Prompt) {
	ed.rprompt = rprompt
}

// startReadLine prepares the terminal for the editor.
func (ed *editor) startReadLine() error {
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
func (ed *editor) finishReadLine() error {
	// After-readline hooks should be called before most teardown happens as
	// they can cause the editor to refresh.
	for _, f := range ed.afterReadline {
		f(ed.buffer)
	}

	ed.activeMutex.Lock()
	defer ed.activeMutex.Unlock()
	ed.active = false

	// Refresh the terminal for the last time in a clean-ish state.
	ed.SetModeInsert()
	ed.tips = nil
	ed.dot = len(ed.buffer)
	if !ed.RpromptPersistent {
		ed.rpromptContent = nil
	}
	errRefresh := ed.refresh(false, false)
	ed.mode.Teardown()
	ed.out.WriteString("\n")
	ed.writer.ResetCurrentBuffer()

	ed.reader.Stop()

	// Restore termios.
	errRestore := ed.restoreTerminal()

	// Reset all of editorState.
	ed.editorState = editorState{}

	return util.Errors(errRefresh, errRestore)
}

// ReadLine reads a line interactively.
func (ed *editor) ReadLine() (string, error) {
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

	ed.SetModeInsert()

	// Find external commands asynchronously, so that slow I/O won't block the
	// editor.
	isExternalCh := make(chan map[string]bool, 1)
	go getIsExternal(isExternalCh)

	ed.reader.Start()

	fullRefresh := false

	for _, f := range ed.beforeReadline {
		f()
	}

	ed.promptContent = ed.prompt.Last()
	ed.rpromptContent = ed.rprompt.Last()
	fresh := true
MainLoop:
	for {
		ed.prompt.Update(fresh)
		ed.rprompt.Update(fresh)
		fresh = false

	refresh:
		err := ed.refresh(fullRefresh, true)
		fullRefresh = false
		if err != nil {
			return "", err
		}

		ed.tips = nil

		select {
		case ed.promptContent = <-ed.prompt.Chan():
			logger.Println("prompt fetched late")
			goto refresh
		case ed.rpromptContent = <-ed.rprompt.Chan():
			logger.Println("rprompt fetched late")
			goto refresh
		case m := <-isExternalCh:
			ed.isExternal = m
			goto refresh
		case sig := <-ed.sigs:
			// TODO(xiaq): Maybe support customizable handling of signals
			switch sig {
			case syscall.SIGHUP:
				return "", io.EOF
			case syscall.SIGINT:
				// Start over
				ed.mode.Teardown()
				ed.editorState = editorState{
					restoreTerminal: ed.restoreTerminal,
					isExternal:      ed.isExternal,
				}
				ed.SetModeInsert()
				fresh = true
				continue MainLoop
			case sys.SIGWINCH:
				fullRefresh = true
				continue MainLoop
			default:
				ed.AddTip("ignored signal %s", sig)
			}
		case event := <-ed.reader.EventChan():
			switch event := event.(type) {
			case tty.NonfatalErrorEvent:
				ed.Notify("error when reading terminal: %v", event.Err)
			case tty.FatalErrorEvent:
				ed.Notify("fatal error when reading terminal: %v", event.Err)
				return "", event.Err
			case tty.MouseEvent:
				ed.AddTip("mouse: %+v", event)
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
				ed.InsertAtDot(topaste)
			case tty.RawRune:
				insertRaw(ed, rune(event))
			case tty.KeyEvent:
				k := ui.Key(event)
			lookupKey:
				fn := ed.mode.Binding(k)
				if fn == nil {
					ed.AddTip("Unbound and no default binding: %s", k)
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
func getIsExternal(result chan<- map[string]bool) {
	isExternal := make(map[string]bool)
	eval.EachExternal(func(name string) {
		isExternal[name] = true
	})
	result <- isExternal
}
