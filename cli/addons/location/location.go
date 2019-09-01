// Package location implements an addon that supports viewing location history
// and changing to a selected directory.
package location

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/combobox"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

// Config is the configuration to start the location history feature.
type Config struct {
	// Binding is the key binding.
	Binding el.Handler
	// Store provides the directory history and the function to change directory.
	Store Store
}

// Store defines the interface for interacting with the directory history.
type Store interface {
	Dirs() ([]storedefs.Dir, error)
	Chdir(dir string) error
}

// Start starts the directory history feature.
func Start(app *cli.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no dir history store")
		return
	}
	dirs, err := cfg.Store.Dirs()
	if err != nil {
		app.Notify("db error: " + err.Error())
		return
	}

	w := combobox.Widget{}
	w.CodeArea.Prompt = layout.ModePrompt("LOCATION", true)
	w.ListBox.OverlayHandler = cfg.Binding
	w.OnFilter = func(p string) {
		w.ListBox.MutateListboxState(func(s *listbox.State) {
			*s = listbox.MakeState(filter(dirs, p), false)
		})
	}
	w.ListBox.OnAccept = func(it listbox.Items, i int) {
		err := cfg.Store.Chdir(it.(items)[i].Path)
		if err != nil {
			app.Notify(err.Error())
		}
		app.MutateAppState(func(s *cli.State) { s.Listing = nil })
	}
	app.MutateAppState(func(s *cli.State) { s.Listing = &w })
}

func filter(dirs []storedefs.Dir, p string) items {
	if p == "" {
		return dirs
	}
	var entries []storedefs.Dir
	for _, dir := range dirs {
		if strings.Contains(dir.Path, p) {
			entries = append(entries, dir)
		}
	}
	return items(entries)
}

type items []storedefs.Dir

func (it items) Show(i int) styled.Text {
	return styled.Plain(fmt.Sprintf("%3.0f %s", it[i].Score, it[i].Path))
}

func (it items) Len() int { return len(it) }
