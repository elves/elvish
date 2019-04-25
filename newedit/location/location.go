package location

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/newedit/listing"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

// Mode represents the location mode. It implements the clitypes.Mode interface by
// embedding a *listing.Mode.
type Mode struct {
	*listing.Mode
	KeyHandler func(ui.Key) clitypes.HandlerAction
	Cd         func(string) error
}

// Start starts the location mode.
func (m *Mode) Start(dirs []storedefs.Dir) {
	m.Mode.Start(listing.StartConfig{
		Name:       "LOCATION",
		KeyHandler: m.KeyHandler,
		ItemsGetter: func(p string) listing.Items {
			return getItems(dirs, p, m.Cd)
		},
		StartFilter: true,
	})
}

func getItems(dirs []storedefs.Dir, p string, cd func(string) error) items {
	var entries []storedefs.Dir
	for _, dir := range dirs {
		if strings.Contains(dir.Path, p) {
			entries = append(entries, dir)
		}
	}
	return items{entries, cd}
}

// A slice of entries plus a cd callback, implementing the listing.Items
// interface.
type items struct {
	entries []storedefs.Dir
	cd      func(string) error
}

func (it items) Len() int {
	return len(it.entries)
}

func (it items) Show(i int) styled.Text {
	return styled.Plain(
		fmt.Sprintf("%3.0f %s", it.entries[i].Score, it.entries[i].Path))
}

func (it items) Accept(i int, st *clitypes.State) {
	err := it.cd(it.entries[i].Path)
	if err != nil {
		st.AddNote(err.Error())
	}
}
