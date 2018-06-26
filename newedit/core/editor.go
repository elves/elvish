package core

import (
	"os"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/loop"
)

type Editor struct {
	loop   *loop.Loop
	tty    TTY
	reader tty.Reader
	writer tty.Writer

	config *Config
	state  *State
}

func NewEditor(r tty.Reader, w tty.Writer, t TTY) *Editor {
	lp := loop.New()
	ed := &Editor{lp, t, r, w, newConfig(), newState()}
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
		action := st.Mode.HandleKey(ui.Key(event), st)
		switch action {
		case CommitCode:
			return st.Code, true
		}
	}
	return "", false
}

func (ed *Editor) redraw(flag loop.RedrawFlag) {
	redraw(ed.state, ed.config, ed.writer, ed.tty.Size, flag)
}

func redraw(st *State, cfg *Config, w tty.Writer, g func() (h, w int), flag loop.RedrawFlag) {
	final := flag&loop.FinalRedraw != 0
	if final {
		st = st.final()
	}

	_, width := g()

	bufNotes, bufMain := render(st, cfg.RenderConfig, width)
	if final {
		bufMain.Newline()
		bufMain.SetDot(bufMain.Cursor())
	}

	w.CommitBuffer(bufNotes, bufMain, flag&loop.FullRedraw != 0)
}

func (ed *Editor) Read() (string, error) {
	restore, err := ed.tty.Setup()
	if err != nil {
		return "", err
	}
	defer restore()

	ed.reader.Start()
	defer ed.reader.Stop()
	go func() {
		for event := range ed.reader.EventChan() {
			ed.loop.Input(event)
		}
	}()

	return ed.loop.Run()
}

func NewStdEditor() *Editor {
	return NewEditor(
		tty.NewReader(os.Stdin), tty.NewWriter(os.Stdout),
		newTTY(os.Stdin, os.Stdout))
}
