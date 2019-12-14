package listbox

import (
	"fmt"

	"github.com/elves/elvish/ui"
)

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Items    Items
	Selected int
	First    int
	Height   int
}

// Items is an interface for accessing multiple items.
type Items interface {
	// Show renders the item at the given zero-based index.
	Show(i int) ui.Text
	// Len returns the number of items.
	Len() int
}

// TestItems is an implementation of Items useful for testing.
type TestItems struct {
	Prefix string
	Style  ui.Styling
	NItems int
}

// Show returns a plain text consisting of the prefix and i. If the prefix is
// empty, it defaults to "item ".
func (it TestItems) Show(i int) ui.Text {
	prefix := it.Prefix
	if prefix == "" {
		prefix = "item "
	}
	return ui.T(fmt.Sprintf("%s%d", prefix, i), it.Style)
}

// Len returns it.NItems.
func (it TestItems) Len() int {
	return it.NItems
}
