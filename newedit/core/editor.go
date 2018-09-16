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

	ConfigMutex sync.RWMutex
	StateMutex  sync.RWMutex

	// Internal states
	prompt, rprompt styled.Text
}

func NewEditor(t TTY, sigs SignalSource) *Editor {
	lp := loop.New()
	ed := &Editor{
		lp, t, sigs,
		&Config{}, &State{},
		sync.RWMutex{}, sync.RWMutex{},
		nil, nil,
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
			ed.StateMutex.Lock()
			*ed.State = State{}
			ed.StateMutex.Unlock()
			ed.triggerPrompts()
		case sys.SIGWINCH:
			ed.Redraw(true)
		}
		return "", false
	case tty.Event:
		switch e := e.(type) {
		case tty.KeyEvent:
			ed.StateMutex.RLock()
			mode := ed.State.Mode
			ed.StateMutex.RUnlock()

			a := getMode(mode).HandleKey(ui.Key(e), ed.State, &ed.StateMutex)

			switch a {
			case CommitCode:
				ed.StateMutex.RLock()
				code := ed.State.Code
				ed.StateMutex.RUnlock()
				return code, true
			}
			ed.triggerPrompts()
		}
		return "", false
	default:
		panic("unreachable")
	}
}

func (ed *Editor) triggerPrompts() {
	ed.ConfigMutex.RLock()
	if ed.Config.RenderConfig.Prompt != nil {
		ed.Config.RenderConfig.Prompt.Trigger()
	}
	if ed.Config.RenderConfig.RPrompt != nil {
		ed.Config.RenderConfig.RPrompt.Trigger()
	}
	defer ed.ConfigMutex.RUnlock()
}

func (ed *Editor) redraw(flag loop.RedrawFlag) {
	ed.StateMutex.RLock()
	defer ed.StateMutex.RUnlock()
	ed.ConfigMutex.RLock()
	defer ed.ConfigMutex.RUnlock()
	redraw(ed.State, ed.Config, ed.tty, ed.tty, flag)
}

func redraw(st *State, cfg *Config, w Output, sz Sizer, flag loop.RedrawFlag) {
	final := flag&loop.FinalRedraw != 0
	if final {
		st = st.final()
	}

	height, width := sz.Size()

	bufNotes, bufMain := render(st, &cfg.RenderConfig, height, width, final)

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
	defer func() {
		ed.StateMutex.Lock()
		defer ed.StateMutex.Unlock()
		*ed.State = State{}
	}()

	// BeforeReadline and AfterReadline hooks.
	ed.ConfigMutex.RLock()
	funcs := ed.Config.BeforeReadline
	ed.ConfigMutex.RUnlock()
	for _, f := range funcs {
		f()
	}
	defer func() {
		ed.ConfigMutex.RLock()
		funcs := ed.Config.AfterReadline
		ed.ConfigMutex.RUnlock()

		ed.StateMutex.RLock()
		code := ed.State.Code
		ed.StateMutex.RUnlock()

		for _, f := range funcs {
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
