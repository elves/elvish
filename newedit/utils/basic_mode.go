package utils

import (
	"unicode/utf8"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/types"
)

// BasicMode is a basic Mode implementation.
type BasicMode struct{}

// ModeLine returns nil.
func (BasicMode) ModeLine() ui.Renderer {
	return nil
}

// ModeRenderFlag returns 0.
func (BasicMode) ModeRenderFlag() types.ModeRenderFlag {
	return 0
}

// HandleEvent uses BasicHandler to handle the event.
func (BasicMode) HandleEvent(e tty.Event, st *types.State) types.HandlerAction {
	return BasicHandler(e, st)
}

// BasicHandler is a basic implementation of an event handler. It is used in
// BasicMode.HandleEvent, but can also be used in other modes as a fallback
// handler.
func BasicHandler(e tty.Event, st *types.State) types.HandlerAction {
	keyEvent, ok := e.(tty.KeyEvent)
	if !ok {
		return types.NoAction
	}
	k := ui.Key(keyEvent)

	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	switch k {
	case ui.Key{Rune: '\n'}:
		return types.CommitCode
	case ui.Key{Rune: ui.Backspace}:
		beforeDot := st.Raw.Code[:st.Raw.Dot]
		afterDot := st.Raw.Code[st.Raw.Dot:]
		_, chop := utf8.DecodeLastRuneInString(beforeDot)
		st.Raw.Code = beforeDot[:len(beforeDot)-chop] + afterDot
		st.Raw.Dot -= chop
	case ui.Key{Rune: ui.Left}:
		_, skip := utf8.DecodeLastRuneInString(st.Raw.Code[:st.Raw.Dot])
		st.Raw.Dot -= skip
	case ui.Key{Rune: ui.Right}:
		_, skip := utf8.DecodeRuneInString(st.Raw.Code[st.Raw.Dot:])
		st.Raw.Dot += skip
	default:
		if k.Mod == 0 {
			s := string(k.Rune)
			st.Raw.Code += s
			st.Raw.Dot += len(s)
		}
	}
	return types.NoAction
}
