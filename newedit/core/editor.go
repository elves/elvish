package core

import (
	"os"
	"sync"
	"syscall"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/loop"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

type Editor struct {
	// Dependencies
	loop *loop.Loop
	tty  TTY
	sigs SignalSource

	Config *Config
	State  *State

	// Internal states
	prompt, rprompt styled.Text
}

func NewEditor(t TTY, sigs SignalSource) *Editor {
	lp := loop.New()
	ed := &Editor{
		lp, t, sigs, &Config{}, &State{}, nil, nil,
	}
	lp.HandleCb(ed.handle)
	lp.RedrawCb(ed.redraw)
	return ed
}

func (ed *Editor) handle(e loop.Event) (string, bool) {
	switch e := e.(type) {
	case os.Signal:
		switch e {
		case syscall.SIGHUP:
			return "", true
		case syscall.SIGINT:
			ed.State.Reset()
			ed.triggerPrompts()
		case sys.SIGWINCH:
			ed.Redraw(true)
		}
		return "", false
	case tty.Event:
		switch e := e.(type) {
		case tty.KeyEvent:
			action := ed.State.Mode().HandleKey(ui.Key(e), ed.State)

			switch action {
			case CommitCode:
				return ed.State.Code(), true
			}
			ed.triggerPrompts()
		}
		return "", false
	default:
		panic("unreachable")
	}
}

func (ed *Editor) triggerPrompts() {
	ed.Config.triggerPrompts()
}

func (ed *Editor) redraw(flag loop.RedrawFlag) {
	redraw(ed.State, ed.Config, ed.tty, ed.tty, flag)
}

func redraw(s *State, cfg *Config, w Output, sz Sizer, flag loop.RedrawFlag) {
	var rawState *RawState
	final := flag&loop.FinalRedraw != 0
	if final {
		rawState = s.finalize()
	} else {
		rawState = s.CopyRaw()
	}

	height, width := sz.Size()

	bufNotes, bufMain := render(rawState, makeRenderSetup(cfg, height, width))

	w.UpdateBuffer(bufNotes, bufMain, flag&loop.FullRedraw != 0)

	if final {
		w.Newline()
		w.ResetBuffer()
	}
}

func (ed *Editor) ReadCode() (string, error) {
	restore, err := ed.tty.Setup()
	if err != nil {
		return "", err
	}
	defer restore()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Relay input events.
	eventCh := ed.tty.StartInput()
	defer ed.tty.StopInput()
	wg.Add(1)
	go func() {
		for event := range eventCh {
			ed.loop.Input(event)
		}
		wg.Done()
	}()

	if ed.sigs != nil {
		// Relay signals.
		sigCh := ed.sigs.NotifySignals()
		defer ed.sigs.StopSignals()
		wg.Add(1)
		go func() {
			for sig := range sigCh {
				ed.loop.Input(sig)
			}
			wg.Done()
		}()
	}

	// TODO: relay late prompt/rprompt updates.

	// Reset state before returning.
	defer ed.State.Reset()

	// BeforeReadline and AfterReadline hooks.
	for _, f := range ed.Config.BeforeReadline() {
		f()
	}
	defer func() {
		code := ed.State.Code()
		for _, f := range ed.Config.AfterReadline() {
			f(code)
		}
	}()

	return ed.loop.Run()
}

// Like ReadCode, but returns immediately with two channels that will get the
// return values of ReadCode. Useful in tests.
func (ed *Editor) readCodeAsync() (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := ed.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
}

func (ed *Editor) Redraw(full bool) {
	ed.loop.Redraw(full)
}
