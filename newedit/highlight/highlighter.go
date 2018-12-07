package highlight

import (
	"sync"

	"github.com/elves/elvish/styled"
)

const latesBufferSize = 128

// Highlighter is a code highlighter that can deliver results asynchronously.
type Highlighter struct {
	dep   Dep
	state state
	lates chan struct{}
}

type state struct {
	sync.RWMutex
	code       string
	styledCode styled.Text
	errors     []error
}

func NewHighlighter(dep Dep) *Highlighter {
	return &Highlighter{dep, state{}, make(chan struct{}, latesBufferSize)}
}

// Get returns the highlighted code and static errors found in the code.
func (hl *Highlighter) Get(code string) (styled.Text, []error) {
	hl.state.RLock()
	if code == hl.state.code {
		hl.state.RUnlock()
		return hl.state.styledCode, hl.state.errors
	}
	hl.state.RUnlock()

	lateCb := func(styledCode styled.Text) {
		hl.state.Lock()
		hl.state.styledCode = styledCode
		hl.state.Unlock()
		hl.lates <- struct{}{}
	}
	return highlight(code, hl.dep, lateCb)
}

// LateUpdates returns a channel for notifying late updates.
func (hl *Highlighter) LateUpdates() <-chan struct{} {
	return hl.lates
}
