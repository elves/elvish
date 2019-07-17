// Package insert is the Elvish-agnostic core of the insert mode. The mode has
// an event handler that handles brackted pasting and implements abbreviation
// expansion.
package insert

import (
	"strings"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/parse"
)

// Mode represents the insert mode, implementing the clitypes.Mode interface.
type Mode struct {
	// Configuration that can be modified concurrently.
	Config Config

	// Internal states.
	inserts string
	paste   pasteStatus
	pastes  []rune
}

type pasteStatus int

const (
	noPaste pasteStatus = iota
	quotePaste
	literalPaste
)

var (
	quotePasteModeLine   = ui.NewModeLineRenderer(" INSERT (pasting, quoted) ", "")
	literalPasteModeLine = ui.NewModeLineRenderer(" INSERT (pasting, literal) ", "")
)

func (m *Mode) initConfig() {
	if m.Config == nil {
		m.Config = DefaultConfig{}
	}
}

// ModeLine returns the modeline.
func (m *Mode) ModeLine() ui.Renderer {
	switch {
	case m.paste == quotePaste:
		return quotePasteModeLine
	case m.paste == literalPaste:
		return literalPasteModeLine
	default:
		return nil
	}
}

// ModeRenderFlag always returns 0.
func (m *Mode) ModeRenderFlag() clitypes.ModeRenderFlag {
	return 0
}

// HandleEvent handles a terminal event. It handles tty.PasteSetting and
// tty.KeyEvent and ignores others.
func (m *Mode) HandleEvent(e term.Event, st *clitypes.State) clitypes.HandlerAction {
	m.initConfig()

	switch e := e.(type) {
	case term.PasteSetting:
		if e {
			m.handlePasteStart()
		} else {
			m.handlePasteEnd(st)
		}
	case term.KeyEvent:
		k := ui.Key(e)
		if m.paste != noPaste {
			m.handleKeyInPaste(k)
			return clitypes.NoAction
		}
		return m.handleKey(k, st)
	}
	return clitypes.NoAction
}

func (m *Mode) handlePasteStart() {
	m.inserts = ""
	if m.Config.QuotePaste() {
		m.paste = quotePaste
	} else {
		m.paste = literalPaste
	}
}

func (m *Mode) handlePasteEnd(st *clitypes.State) {
	text := string(m.pastes)
	if m.paste == quotePaste {
		text = parse.Quote(text)
	}
	st.InsertAtDot(text)
	m.pastes = nil
	m.paste = noPaste
}

func (m *Mode) handleKeyInPaste(k ui.Key) {
	if k.Mod != 0 || k.Rune < 0 {
		// TODO: Notify user of the error, instead of silently dropping
		// function keys.
	} else {
		m.pastes = append(m.pastes, k.Rune)
	}
}

func (m *Mode) handleKey(k ui.Key, st *clitypes.State) clitypes.HandlerAction {
	action := m.Config.HandleKey(k, st)
	if k.Mod != 0 || k.Rune < 0 {
		m.inserts = ""
		return action
	}
	// Expand abbreviations.
	//
	// NOTE: If the user binds a non-function key to do something other than
	// inserting the key itself, the behavior can be confusing.
	m.inserts += string(k.Rune)
	var abbr, full string
	m.Config.IterateAbbr(func(a, f string) {
		if strings.HasSuffix(m.inserts, a) && len(a) > len(abbr) {
			abbr, full = a, f
		}
	})
	if len(abbr) > 0 {
		st.Mutex.Lock()
		defer st.Mutex.Unlock()
		raw := &st.Raw
		// Make sure that abbr is to the left of the dot.
		if strings.HasSuffix(raw.Code[:raw.Dot], abbr) {
			// Replace abbr with full
			beforeAbbr := raw.Code[:raw.Dot-len(abbr)]
			raw.Code = beforeAbbr + full + raw.Code[raw.Dot:]
			raw.Dot = len(beforeAbbr) + len(full)
		}
	}
	return action
}
