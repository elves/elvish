package highlight

import (
	"sync"

	"src.elv.sh/pkg/ui"
)

const latesBufferSize = 128

// Highlighter is a code highlighter that can deliver results asynchronously.
type Highlighter struct {
	cfg   Config
	lates chan struct{}

	cacheMutex sync.Mutex
	cache      cache
}

type cache struct {
	code       string
	styledCode ui.Text
	tips       []ui.Text
}

func NewHighlighter(cfg Config) *Highlighter {
	return &Highlighter{cfg: cfg, lates: make(chan struct{}, latesBufferSize)}
}

// Get returns the highlighted code and static errors found in the code as tips.
func (hl *Highlighter) Get(code string) (ui.Text, []ui.Text) {
	hl.cacheMutex.Lock()
	defer hl.cacheMutex.Unlock()
	if code == hl.cache.code {
		return hl.cache.styledCode, hl.cache.tips
	}

	lateCb := func(styledCode ui.Text) {
		hl.cacheMutex.Lock()
		if hl.cache.code != code {
			// Late result was delivered after code has changed. Unlock and
			// return.
			hl.cacheMutex.Unlock()
			return
		}
		hl.cache.styledCode = styledCode
		// The channel send below might block, so unlock the state first.
		hl.cacheMutex.Unlock()
		hl.lates <- struct{}{}
	}

	styledCode, tips := highlight(code, hl.cfg, lateCb)

	hl.cache = cache{code, styledCode, tips}
	return styledCode, tips
}

// LateUpdates returns a channel for notifying late updates.
func (hl *Highlighter) LateUpdates() <-chan struct{} {
	return hl.lates
}

// InvalidateCache invalidates the cached highlighting result.
func (hl *Highlighter) InvalidateCache() {
	hl.cacheMutex.Lock()
	defer hl.cacheMutex.Unlock()
	hl.cache = cache{}
}
