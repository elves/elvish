package core

import (
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/types"
)

type basicMode struct{}

func (basicMode) ModeLine() ui.Renderer {
	return nil
}

func (basicMode) ModeRenderFlag() types.ModeRenderFlag {
	return 0
}

func (basicMode) HandleKey(k ui.Key, st *types.State) types.HandlerAction {
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

func getMode(m types.Mode) types.Mode {
	if m == nil {
		return basicMode{}
	}
	return m
}
