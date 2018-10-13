// Package insert is the Elvish-agnostic core of the insert mode. The mode has
// an event handler that handles brackted pasting and implements abbreviation
// expansion.
package insert

import (
	"strings"
	"sync"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/editutil"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/parse"
)

// Mode represents the insert mode, implementing the types.Mode interface.
type Mode struct {
	// Function to handle keys.
	KeyHandler func(ui.Key) types.HandlerAction
	// Function that feeds all abbreviation pairs to the callback.
	AbbrIterate func(func(abbr, full string))
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

// Config keeps configurations that can be changed concurrently.
type Config struct {
	Raw   RawConfig
	Mutex sync.RWMutex
}

// QuotePaste returns c.Raw.QuotePaste while r-locking c.Mutex.
func (c *Config) QuotePaste() bool {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.Raw.QuotePaste
}

// RawConfig keeps raw configurations.
type RawConfig struct {
	// Whether to quote the bracketed-pasted text.
	QuotePaste bool
}

var (
	quotePasteModeLine   = ui.NewModeLineRenderer(" INSERT (pasting, quoted) ", "")
	literalPasteModeLine = ui.NewModeLineRenderer(" INSERT (pasting, literal) ", "")
)

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
func (m *Mode) ModeRenderFlag() types.ModeRenderFlag {
	return 0
}

// HandleEvent handles a terminal event. It handles tty.PasteSetting and
// tty.KeyEvent and ignores others.
func (m *Mode) HandleEvent(e tty.Event, st *types.State) types.HandlerAction {
	switch e := e.(type) {
	case tty.PasteSetting:
		if e {
			m.handlePasteStart()
		} else {
			m.handlePasteEnd(st)
		}
	case tty.KeyEvent:
		k := ui.Key(e)
		if m.paste != noPaste {
			m.handleKeyInPaste(k)
			return types.NoAction
		}
		return m.handleKey(k, st)
	}
	return types.NoAction
}

func (m *Mode) handlePasteStart() {
	m.inserts = ""
	if m.Config.QuotePaste() {
		m.paste = quotePaste
	} else {
		m.paste = literalPaste
	}
}

func (m *Mode) handlePasteEnd(st *types.State) {
	text := string(m.pastes)
	if m.paste == quotePaste {
		text = parse.Quote(text)
	}
	insert(st, text)
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

func (m *Mode) handleKey(k ui.Key, st *types.State) types.HandlerAction {
	var action types.HandlerAction
	if m.KeyHandler != nil {
		action = m.KeyHandler(k)
	} else {
		action = editutil.BasicHandler(tty.KeyEvent(k), st)
	}
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
	if m.AbbrIterate != nil {
		m.AbbrIterate(func(a, f string) {
			if strings.HasSuffix(m.inserts, a) && len(a) > len(abbr) {
				abbr, full = a, f
			}
		})
	}
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

func insert(st *types.State, text string) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()
	raw := &st.Raw
	raw.Code = raw.Code[:raw.Dot] + text + raw.Code[raw.Dot:]
	raw.Dot += len(text)
}
