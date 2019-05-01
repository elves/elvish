package histlist

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/histutil"
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
func (m *Mode) Start(cmds []histutil.Entry) {
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
func getItems(cmds []histutil.Entry, p string) items {
	var entries []histutil.Entry
	for _, entry := range cmds {
		if strings.Contains(entry.Text, p) {
			entries = append(entries, entry)
		}
	}
	return entries
}

// A slice of entries, implementing the listing.Items interface.
type items []histutil.Entry

func (it items) Len() int {
	return len(it)
}

func (it items) Show(i int) styled.Text {
	// TODO: The alignment of the index works up to 10000 entries.
	return styled.Plain(fmt.Sprintf("%4d %s", it[i].Seq, it[i].Text))
}

func (it items) Accept(i int, st *clitypes.State) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()
	raw := &st.Raw

	if raw.Code == "" {
		insertAtDot(raw, it[i].Text)
	} else {
		// TODO: This works well when the cursor is at the end, but can be
		// unexpected when the cursor is in the middle.
		insertAtDot(raw, "\n"+it[i].Text)
	}
}

func insertAtDot(raw *clitypes.RawState, text string) {
	// NOTE: This is an duplicate with (*clitypes.State).InsertAtDot, without any
	// locks because we accept RawState.
	raw.Code = raw.Code[:raw.Dot] + text + raw.Code[raw.Dot:]
	raw.Dot += len(text)
}
