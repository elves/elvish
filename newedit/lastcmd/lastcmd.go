package lastcmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/listing"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// Mode represents the lastcmd mode. It implements the clitypes.Mode interface by
// embedding a *listing.Mode.
type Mode struct {
	*listing.Mode
	KeyHandler func(ui.Key) clitypes.HandlerAction
}

// Start starts the lastcmd mode.
func (m *Mode) Start(line string, words []string) {
	m.Mode.Start(listing.StartConfig{
		Name:        "LASTCMD",
		KeyHandler:  m.KeyHandler,
		ItemsGetter: itemsGetter(line, words),
		StartFilter: true,
		AutoAccept:  true,
	})
}

func itemsGetter(line string, words []string) func(string) listing.Items {
	// Build the list of all entries from the line and words. Entries have
	// positive and negative indicies, except for the first entry, which
	// represents the entire line and has no indicies.
	entries := make([]entry, len(words)+1)
	entries[0] = entry{content: line}
	for i, word := range words {
		entries[i+1] = entry{
			posIndex: strconv.Itoa(i),
			negIndex: strconv.Itoa(i - len(words)),
			content:  word,
		}
	}

	return func(p string) listing.Items {
		return filter(entries, p)
	}
}

func filter(allEntries []entry, p string) items {
	var entries []entry
	negFilter := strings.HasPrefix(p, "-")
	for _, entry := range allEntries {
		index := ""
		if negFilter {
			index = entry.negIndex
		} else {
			index = entry.posIndex
		}
		if strings.HasPrefix(index, p) {
			entries = append(entries, entry)
		}
	}
	return items{negFilter, entries}
}

type items struct {
	negFilter bool
	entries   []entry
}

type entry struct {
	posIndex string
	negIndex string
	content  string
}

func (it items) Len() int {
	return len(it.entries)
}

func (it items) Show(i int) styled.Text {
	// NOTE: We now use a hardcoded width of 3 for the index, which will work as
	// long as the command has less than 1000 words (when filter is positive) or
	// 100 words (when filter is negative).
	index := ""
	entry := it.entries[i]
	if it.negFilter {
		index = entry.negIndex
	} else {
		index = entry.posIndex
	}
	return styled.Plain(fmt.Sprintf("%3s %s", index, entry.content))
}

func (it items) Accept(i int, st *clitypes.State) {
	st.InsertAtDot(it.entries[i].content)
}
