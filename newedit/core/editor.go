package core

import (
	"sync"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/loop"
)

type Editor struct {
	loop *loop.Loop
	tty  TTY
	sigs SignalSource

	config *Config
	state  *State
}

func NewEditor(t TTY, sigs SignalSource) *Editor {
	lp := loop.New()
	ed := &Editor{lp, t, sigs, &Config{}, &State{}}
	lp.HandleCb(ed.handle)
	lp.RedrawCb(ed.redraw)
	return ed
}

func (ed *Editor) handle(e loop.Event) (string, bool) {
	return handle(ed.state, e.(tty.Event))
}

func handle(st *State, event tty.Event) (string, bool) {
	switch event := event.(type) {
	case tty.KeyEvent:
		action := getMode(st.Mode).HandleKey(ui.Key(event), st)
		switch action {
		case CommitCode:
			return st.Code, true
		}
	}
	return "", false
}

func (ed *Editor) redraw(flag loop.RedrawFlag) {
	redraw(ed.state, ed.config, ed.tty, ed.tty, flag)
}

func redraw(st *State, cfg *Config, w Writer, sz Sizer, flag loop.RedrawFlag) {
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
	eventCh := ed.tty.StartRead()
	defer func() {
		ed.tty.StopRead()
		wg.Wait()
	}()
	wg.Add(1)
	go func() {
		for event := range eventCh {
			ed.loop.Input(event)
		}
		wg.Done()
	}()

	for _, f := range ed.config.BeforeReadline {
		f()
	}
	defer func() {
		for _, f := range ed.config.AfterReadline {
			f(ed.state.Code)
		}
	}()

	return ed.loop.Run()
}

func (ed *Editor) Redraw(full bool) {
	ed.loop.Redraw(full)
}
