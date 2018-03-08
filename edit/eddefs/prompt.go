package eddefs

import "github.com/elves/elvish/edit/ui"

// Prompt is the interface for a general prompt.
type Prompt interface {
	// Chan returns a prompt on which the content of the prompt is made
	// available.
	Chan() <-chan []*ui.Styled
	// Update signifies that the prompt should be updated.
	Update()
	// ForceUpdate signifies that the prompt should be updated and requires a
	// prompt to be returned immediately.
	ForceUpdate() []*ui.Styled
	// Close releases resources associated with the prompt.
	Close() error
}
