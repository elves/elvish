package core

import (
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
)

// Mode is an editor mode; it handles keys and can affect the current UI.
type Mode interface {
	// ModeLine returns a Renderer for the modeline. It may return nil, in which
	// case the modeline is hidden.
	ModeLine() ui.Renderer
	// ModeRenderFlag returns flags that can affect the UI.
	ModeRenderFlag() ModeRenderFlag
	// HandleKey handles a key event; its return value can affect the editor
	// lifecycle.
	HandleKey(ui.Key, *State) HandlerAction
}

// Lister is an optional interface that modes can implement. If a mode
// implements this interface, the result of this method is shown in the listing
// area.
type Lister interface {
	List(maxHeight int) ui.Renderer
}

// ModeRenderFlag is a bitmask for configuring the rendering behavior of modes.
type ModeRenderFlag uint

// Bits for ModeRenderFlag.
const (
	// Place the cursor on the mode line (instead of the code area).
	CursorOnModeLine = 1 << iota
	// Redraw the modeline after List. Has not effect if the mode does not
	// implement Lister.
	RedrawModeLineAfterList
)

// HandlerAction is used as the return code of Mode.HandleKey and can affect the
// editor lifecycle.
type HandlerAction int

const (
	// NoAction is the default value of HandlerAction, which enacts no effect on
	// the editor lifecycle.
	NoAction HandlerAction = iota
	// CommitCode will cause the editor's ReadCode function to return with the
	// current code.
	CommitCode
	// CommitEOF
)

type basicMode struct{}

func (basicMode) ModeLine() ui.Renderer {
	return nil
}

func (basicMode) ModeRenderFlag() ModeRenderFlag {
	return 0
}

func (basicMode) HandleKey(k ui.Key, st *State) HandlerAction {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	switch k {
	case ui.Key{Rune: '\n'}:
		return CommitCode
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
	return NoAction
}
