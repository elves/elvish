package highlight

import (
	"github.com/elves/elvish/styled"
)

// Highlighter is a code highlighter that can deliver results asynchronously.
type Highlighter struct {
	dep Dep
}

func NewHighlighter(dep Dep) *Highlighter {
	return &Highlighter{dep}
}

// Get returns the highlighted code and static errors found in the code.
func (hl *Highlighter) Get(code string) (styled.Text, []error) {
	return highlight(code, hl.dep)
}

// LateUpdates returns a channel for notifying late updates.
func (hl *Highlighter) LateUpdates() <-chan struct{} {
	return nil
}
