package etk

import (
	"io"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/ui"
)

type RunCfg struct {
	TTY       cli.TTY
	Frame     *eval.Frame
	MaxHeight int
	Justify   Justify
	ContextFn func(Context)
}

type Justify uint8

const (
	NoJustify Justify = iota
	JustifyTop
	JustifyCenter
	JustifyBottom
)

func Run(f Comp, cfg RunCfg) (vals.Map, error) {
	tty, fm := cfg.TTY, cfg.Frame
	restore, err := tty.Setup()
	if err != nil {
		return nil, err
	}
	defer restore()

	// Start reading events.
	eventCh := make(chan term.Event)
	go func() {
		for {
			event, err := tty.ReadEvent()
			if err != nil {
				if err == term.ErrStopped {
					return
				}
				// TODO: Report error in notification
			}
			eventCh <- event
		}
	}()
	defer tty.CloseReader()

	sc := Stateful(fm, f)
	defer sc.Finish()

	if cfg.ContextFn != nil {
		cfg.ContextFn(Context{sc.g, nil})
	}

	for {
		// TODO: Consume any
		// Render.
		h, w := tty.Size()
		if cfg.MaxHeight > 0 {
			h = min(h, cfg.MaxHeight)
		}
		buf := justify(sc.Render(w, h), h, cfg.Justify)
		msg := sc.MergeAndPopMsgs(w)
		tty.UpdateBuffer(msg, buf, false /*true*/)

		select {
		case event := <-eventCh:
			reaction := sc.React(event)
			if reaction == Finish || reaction == FinishEOF {
				h, w := tty.Size()
				buf := sc.Render(w, h)
				msg := sc.MergeAndPopMsgs(w)
				// Render the final view with a trailing newline. This operation
				// is quite subtle with the term.Buffer API.
				buf.ExtendDown(term.NewBufferBuilder(w).Buffer(), true)
				tty.UpdateBuffer(msg, buf, false)
				if reaction == FinishEOF {
					return sc.g.state, io.EOF
				} else {
					return sc.g.state, nil
				}
			}
		case <-sc.g.refreshCh:
			sc.Refresh()
		}
	}
}

func justify(buf *term.Buffer, h int, j Justify) *term.Buffer {
	w := buf.Width
	padding := h - len(buf.Lines)
	if padding == 0 {
		return buf
	}
	switch j {
	case JustifyTop:
		buf.ExtendDown(makeEmptyLines(w, padding), false)
	case JustifyCenter:
		topPadding := padding / 2
		if topPadding > 0 {
			buf0 := makeEmptyLines(w, topPadding)
			buf0.ExtendDown(buf, true)
			buf = buf0
		}
		buf.ExtendDown(makeEmptyLines(w, padding-topPadding), false)
	case JustifyBottom:
		buf0 := makeEmptyLines(w, padding)
		buf0.ExtendDown(buf, true)
		buf = buf0
	}
	return buf
}

func makeEmptyLines(w, h int) *term.Buffer {
	buf := term.NewBufferBuilder(w)
	for i := 1; i < h; i++ {
		buf.Newline()
	}
	return buf.Buffer()
}

func Stateful(fm *eval.Frame, f Comp) *StatefulComp {
	g := makeGlobalContext(fm)
	view, react := f(Context{g, nil})
	return &StatefulComp{g, f, view, react}
}

type StatefulComp struct {
	g *globalContext

	f     Comp
	view  View
	react React
}

func (sc *StatefulComp) Render(width, height int) *term.Buffer {
	return sc.view.Render(width, height)
}

func (sc *StatefulComp) React(event term.Event) Reaction {
	reaction := func() Reaction {
		sc.g.batchMutex.Lock()
		defer sc.g.batchMutex.Unlock()
		return sc.react(event)
	}()
	sc.Refresh()
	return reaction
}

func (sc *StatefulComp) Refresh() {
	sc.g.batchMutex.Lock()
	defer sc.g.batchMutex.Unlock()
	// Exhaust the refreshCh, in case superfluous refreshes were requested, such
	// as from an event handler.
	//
	// TODO: Is this the correct place to mitigate superfluous refreshes?
	select {
	case <-sc.g.refreshCh:
	default:
	}
	sc.view, sc.react = sc.f(Context{sc.g, nil})
}

func (sc *StatefulComp) RefreshIfRequested() {
	select {
	case <-sc.g.refreshCh:
		sc.Refresh()
	default:
	}
}

func (sc *StatefulComp) WaitRefresh() {
	<-sc.g.refreshCh
	sc.Refresh()
}

func (sc *StatefulComp) MergeAndPopMsgs(width int) ui.Text {
	msgs := sc.g.PopMsgs()
	if len(msgs) == 0 {
		return nil
	}
	var tb ui.TextBuilder
	for i, msg := range msgs {
		if i > 0 {
			tb.WriteText(ui.T("\n"))
		}
		tb.WriteText(msg)
	}
	return tb.Text()
}

func (sc *StatefulComp) Finish() { close(sc.g.finishCh) }
