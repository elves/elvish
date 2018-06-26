package core

import (
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
)

// Mode is an editor mode.
type Mode interface {
	ModeLine() ui.Renderer
	ModeRenderFlag() ModeRenderFlag
	HandleKey(ui.Key, *State) HandlerAction
	// Teardown()
}

// An optional interface that modes can implement. If a mode implements this
// interface, the result of this method is shown in the listing area.
type Lister interface {
	List(maxHeight int) ui.Renderer
}

// Bitmask for configuring the rendering behavior of modes.
type ModeRenderFlag uint

// Bits for ModeRenderFlag.
const (
	// Place the cursor on the mode line (instead of the code area).
	CursorOnModeLine = 1 << iota
	// Redraw the modeline after List. Has not effect if the mode does not
	// implement Lister.
	RedrawModeLineAfterList
)

type HandlerAction int

const (
	NoAction HandlerAction = iota
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
	switch k {
	case ui.Key{Rune: '\n'}:
		return CommitCode
	case ui.Key{Rune: ui.Backspace}:
		beforeDot := st.CodeBeforeDot()
		_, chop := utf8.DecodeLastRuneInString(beforeDot)
		st.Code = beforeDot[:len(beforeDot)-chop] + st.CodeAfterDot()
		st.Dot -= chop
	case ui.Key{Rune: ui.Left}:
		_, skip := utf8.DecodeLastRuneInString(st.CodeBeforeDot())
		st.Dot -= skip
	case ui.Key{Rune: ui.Right}:
		_, skip := utf8.DecodeRuneInString(st.CodeAfterDot())
		st.Dot += skip
	default:
		if k.Mod == 0 {
			s := string(k.Rune)
			st.Code += s
			st.Dot += len(s)
		}
	}
	return NoAction
}
