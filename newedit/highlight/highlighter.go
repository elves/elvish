package highlight

import (
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
)

// Highlighter is a code highlighter that can deliver results asynchronously.
type Highlighter struct {
	Check      func(n *parse.Chunk) error
	HasCommand func(name string) bool
}

// Get returns the highlighted code and static errors found in the code.
func (hl Highlighter) Get(code string) (styled.Text, []error) {
	return highlight(code, Dep{hl.Check, hl.HasCommand})
}

// LateUpdates returns a channel for notifying late updates.
func (hl Highlighter) LateUpdates() <-chan struct{} {
	return nil
}
