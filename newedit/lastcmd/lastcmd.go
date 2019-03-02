package lastcmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/styled"
)

func StartConfig(line string, words []string) listing.StartConfig {
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

	return listing.StartConfig{
		Name: "LASTCMD",
		ItemsGetter: func(p string) listing.Items {
			return filter(entries, p)
		},
		// TODO: Uncomment
		// AutoAccept: true,
		// StartFiltering: true,
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
	return styled.Unstyled(fmt.Sprintf("%3s %s", index, entry.content))
}
