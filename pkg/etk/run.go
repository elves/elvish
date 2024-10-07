package etk

import (
	"io"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

func Run(tty cli.TTY, fm *eval.Frame, f Comp) (vals.Map, error) {
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

	for {
		// Render.
		h, w := tty.Size()
		buf := sc.Render(w, h)
		msgBuf := sc.RenderAndPopMsgs(w)
		tty.UpdateBuffer(msgBuf, buf, false /*true*/)

		select {
		case event := <-eventCh:
			reaction := sc.Handle(event)
			if reaction == Finish || reaction == FinishEOF {
				h, w := tty.Size()
				buf := sc.Render(w, h)
				msgBuf := sc.RenderAndPopMsgs(w)
				// Render the final view with a trailing newline. This operation
				// is quite subtle with the term.Buffer API.
				buf.Extend(term.NewBufferBuilder(w).Buffer(), true)
				tty.UpdateBuffer(msgBuf, buf, false)
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

func (sc *StatefulComp) Handle(event term.Event) Reaction {
	reaction := sc.callReact(event)
	sc.Refresh()
	return reaction
}

func (sc *StatefulComp) callReact(event term.Event) Reaction {
	sc.g.batchMutex.Lock()
	defer sc.g.batchMutex.Unlock()
	return sc.react(event)
}

func (sc *StatefulComp) Refresh() {
	sc.g.batchMutex.Lock()
	defer sc.g.batchMutex.Unlock()
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

func (sc *StatefulComp) RenderAndPopMsgs(width int) *term.Buffer {
	msgs := sc.g.PopMsgs()
	if len(msgs) == 0 {
		return nil
	}
	bb := term.NewBufferBuilder(width)
	for i, msg := range msgs {
		if i > 0 {
			bb.Newline()
		}
		bb.WriteStyled(msg)
	}
	return bb.Buffer()
}

func (sc *StatefulComp) Finish() { close(sc.g.finishCh) }
