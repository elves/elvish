package eddefs

import "github.com/elves/elvish/edit/ui"

// Prompt is the interface for a general prompt.
type Prompt interface {
	// Chan returns a prompt on which the content of the prompt is made
	// available.
	Chan() <-chan []*ui.Styled
	// Update signifies that the prompt should be updated.
	Update(force bool)
	// Last returns the last content written to Chan. If Chan was never written,
	// it should return some content representing an unknown prompt.
	Last() []*ui.Styled
	// Close releases resources associated with the prompt.
	Close() error
}
