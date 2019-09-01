package listbox

import (
	"fmt"

	"github.com/elves/elvish/styled"
)

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Items    Items
	Selected int
	First    int
}

// MakeState makes a new State.
func MakeState(it Items, selectLast bool) State {
	selected := 0
	if selectLast {
		selected = it.Len() - 1
	}
	return State{it, selected, 0}
}

// Items is an interface for accessing multiple items.
type Items interface {
	// Show renders the item at the given zero-based index.
	Show(i int) styled.Text
	// Len returns the number of items.
	Len() int
}

// TestItems is an implementation of Items useful for testing.
type TestItems struct {
	Prefix string
	NItems int
}

// Show returns a plain text consisting of the prefix and i. If the prefix is
// empty, it defaults to "item ".
func (it TestItems) Show(i int) styled.Text {
	prefix := it.Prefix
	if prefix == "" {
		prefix = "item "
	}
	return styled.Plain(fmt.Sprintf("%s%d", prefix, i))
}

// Len returns it.NItems.
func (it TestItems) Len() int {
	return it.NItems
}
