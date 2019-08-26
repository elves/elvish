package location

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/combobox"
	"github.com/elves/elvish/cli/layout"
	"github.com/elves/elvish/cli/listbox"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

type Config struct {
	Binding clitypes.Handler
	Store   Store
}

// Store defines the interface for interacting with the directory history.
type Store interface {
	Dirs() ([]storedefs.Dir, error)
	Chdir(dir string) error
}

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
	}
	app.MutateAppState(func(s *cli.State) { s.Listing = &w })
}

func filter(dirs []storedefs.Dir, p string) items {
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
