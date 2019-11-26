package highlight

import (
	"sync"

	"github.com/elves/elvish/styled"
)

const latesBufferSize = 128

// Highlighter is a code highlighter that can deliver results asynchronously.
type Highlighter struct {
	cfg   Config
	state state
	lates chan styled.Text
}

type state struct {
	sync.Mutex
	code       string
	styledCode styled.Text
	errors     []error
}

func NewHighlighter(cfg Config) *Highlighter {
	return &Highlighter{cfg, state{}, make(chan styled.Text, latesBufferSize)}
}

// Get returns the highlighted code and static errors found in the code.
func (hl *Highlighter) Get(code string) (styled.Text, []error) {
	hl.state.Lock()
	defer hl.state.Unlock()
	if code == hl.state.code {
		return hl.state.styledCode, hl.state.errors
	}

	lateCb := func(styledCode styled.Text) {
		hl.state.Lock()
		if hl.state.code != code {
			// Late result was delivered after code has changed. Unlock and
			// return.
			hl.state.Unlock()
			return
		}
		hl.state.styledCode = styledCode
		// The channel send below might block, so unlock the state first.
		hl.state.Unlock()
		hl.lates <- styledCode
	}

	styledCode, errors := highlight(code, hl.cfg, lateCb)

	hl.state.code = code
	hl.state.styledCode = styledCode
	hl.state.errors = errors
	return styledCode, errors
}

// LateUpdates returns a channel for notifying late updates.
func (hl *Highlighter) LateUpdates() <-chan styled.Text {
	return hl.lates
}
