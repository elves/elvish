package histlist

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/styled"
)

// Mode represents the histlist mode. It implements the clitypes.Mode interface by
// embedding a *listing.Mode.
type Mode struct {
	*listing.Mode
	KeyHandler func(ui.Key) clitypes.HandlerAction
}

// Start starts the histlist mode.
func (m *Mode) Start(cmds []string) {
	m.Mode.Start(listing.StartConfig{
		Name:       "HISTLIST",
		KeyHandler: m.KeyHandler,
		ItemsGetter: func(p string) listing.Items {
			return getItems(cmds, p)
		},
		StartFilter: true,
	})
}

// Given all commands, and a pattern, returning all matching entries.
func getItems(cmds []string, p string) items {
	// TODO: Show the real in-storage IDs of cmds, not their in-memory indicies.
	var entries []entry
	for i, line := range cmds {
		if strings.Contains(line, p) {
			entries = append(entries, entry{line, i})
		}
	}
	return entries
}

// A slice of entries, implementing the listing.Items interface.
type items []entry

// An entry to show, which is just a line plus its index.
type entry struct {
	content string
	index   int
}

func (it items) Len() int {
	return len(it)
}

func (it items) Show(i int) styled.Text {
	// TODO: The alignment of the index works up to 10000 entries.
	return styled.Plain(fmt.Sprintf("%4d %s", it[i].index+1, it[i].content))
}

func (it items) Accept(i int, st *clitypes.State) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()
	raw := &st.Raw

	if raw.Code == "" {
		insertAtDot(raw, it[i].content)
	} else {
		// TODO: This works well when the cursor is at the end, but can be
		// unexpected when the cursor is in the middle.
		insertAtDot(raw, "\n"+it[i].content)
	}
}

func insertAtDot(raw *clitypes.RawState, text string) {
	// NOTE: This is an duplicate with (*clitypes.State).InsertAtDot, without any
	// locks because we accept RawState.
	raw.Code = raw.Code[:raw.Dot] + text + raw.Code[raw.Dot:]
	raw.Dot += len(text)
}
