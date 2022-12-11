package highlight

import (
	"sync"

	"src.elv.sh/pkg/ui"
)

const latesBufferSize = 128

// Highlighter is a code highlighter that can deliver results asynchronously.
type Highlighter struct {
	cfg   Config
	state state
	lates chan struct{}
}

type state struct {
	sync.Mutex
	code       string
	styledCode ui.Text
	tips       []ui.Text
}

func NewHighlighter(cfg Config) *Highlighter {
	return &Highlighter{cfg, state{}, make(chan struct{}, latesBufferSize)}
}

// Get returns the highlighted code and static errors found in the code as tips.
func (hl *Highlighter) Get(code string) (ui.Text, []ui.Text) {
	hl.state.Lock()
	defer hl.state.Unlock()
	if code == hl.state.code {
		return hl.state.styledCode, hl.state.tips
	}

	lateCb := func(styledCode ui.Text) {
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
		hl.lates <- struct{}{}
	}

	styledCode, errors := highlight(code, hl.cfg, lateCb)
	var tips []ui.Text
	if len(errors) > 0 {
		tips = make([]ui.Text, len(errors))
		for i, err := range errors {
			tips[i] = ui.T(err.Error())
		}
	}

	hl.state.code = code
	hl.state.styledCode = styledCode
	hl.state.tips = tips
	return styledCode, tips
}

// LateUpdates returns a channel for notifying late updates.
func (hl *Highlighter) LateUpdates() <-chan struct{} {
	return hl.lates
}
