package eddefs

import (
	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// Editor is the interface for the Elvish line editor.
type Editor interface {
	// ReadLine reads a line interactively.
	ReadLine() (string, error)
	// Close releases resources used by the editor.
	Close()

	// Evaler returns the Evaler associated with the Editor.
	Evaler() *eval.Evaler
	// Daemon returns the daemon client associated with the Editor.
	Daemon() *daemon.Client

	// Buffer returns the current content and dot position of the buffer.
	Buffer() (string, int)
	// SetBuffer sets the current content and dot position of the buffer.
	SetBuffer(buffer string, dot int)
	// ParsedBuffer returns the node from parsing the buffer.
	ParsedBuffer() *parse.Chunk
	// InsertAtDot inserts text at the dot and moves the dot after it.
	InsertAtDot(text string)

	// SetPrompt sets the prompt of the editor.
	SetPrompt(prompt Prompt)
	// SetPrompt sets the rprompt of the editor.
	SetRPrompt(rprompt Prompt)

	// SetMode sets the current mode of the Editor.
	SetMode(m Mode)
	// SetModeInsert sets the current mode of the Editor to insert mode.
	SetModeInsert()
	// SetModeListing sets the current mode of the Editor to listing mode with
	// the supplied binding and provider.
	SetModeListing(b BindingMap, p ListingProvider)
	// RefreshListing refreshes the listing mode, recalculating the listing
	// items. It is useful when the underlying listing provider has been
	// changed. If the editor is not in listing mode, it does nothing.
	RefreshListing()

	// AddTip adds a message to the tip area.
	AddTip(format string, args ...interface{})
	// Notify writes out a message in a way that does not interrupt the editor
	// display. When the editor is not active, it simply writes the message to
	// the terminal. When the editor is active, it appends the message to the
	// notification queue, which will be written out during the update cycle. It
	// can be safely used concurrently.
	Notify(format string, args ...interface{})

	// LastKey returns the last key received from the user. It is useful mainly
	// in keybindings.
	LastKey() ui.Key

	// SetAction sets the action to execute after the key binding has finished.
	SetAction(a Action)

	// AddAfterReadline adds a hook function that runs after readline ends.
	AddAfterReadline(func(string))
}
