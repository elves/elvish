package etk

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

func Stateful(fm *eval.Frame, f Comp) *StatefulComp {
	g := &globalContext{
		vals.EmptyMap, make(chan struct{}, 1), make(chan struct{}), fm, nil}
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
	reaction := sc.react(event)
	sc.Refresh()
	return reaction
}

func (sc *StatefulComp) Refresh() {
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
	if len(sc.g.msgs) == 0 {
		return nil
	}
	bb := term.NewBufferBuilder(width)
	for i, msg := range sc.g.msgs {
		if i > 0 {
			bb.Newline()
		}
		bb.WriteStyled(msg)
	}
	sc.g.msgs = nil
	return bb.Buffer()
}

func (sc *StatefulComp) Finish() { close(sc.g.finishCh) }
