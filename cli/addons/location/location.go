package location

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/combobox"
	"github.com/elves/elvish/cli/layout"
	"github.com/elves/elvish/cli/listbox"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

type Config struct {
	Binding clitypes.Handler
	Store   DirStore
}

// DirStore defines the interface for interacting with the directory history.
type DirStore interface {
	Dirs() ([]storedefs.Dir, error)
	Chdir(dir string) error
}

func Start(app *clicore.App, cfg Config) {
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
	w.CodeArea.State.Prompt = layout.ModePrompt("LOCATION", true)
	w.ListBox.OverlayHandler = cfg.Binding
	w.OnFilter = func(p string) {
		w.ListBox.MutateListboxState(func(s *listbox.State) {
			itemer, n := filter(dirs, p)
			*s = listbox.MakeState(itemer, n, false)
		})
	}
	w.ListBox.OnAccept = func(i int) {
		itemer := w.ListBox.CopyListboxState().Itemer.(itemer)
		err := cfg.Store.Chdir(itemer[i].Path)
		if err != nil {
			app.Notify(err.Error())
		}
	}
	app.MutateAppState(func(s *clicore.State) { s.Listing = &w })
}

func filter(dirs []storedefs.Dir, p string) (itemer, int) {
	var entries []storedefs.Dir
	for _, dir := range dirs {
		if strings.Contains(dir.Path, p) {
			entries = append(entries, dir)
		}
	}
	return itemer(entries), len(entries)
}

type itemer []storedefs.Dir

func (it itemer) Item(i int) styled.Text {
	return styled.Plain(fmt.Sprintf("%3.0f %s", it[i].Score, it[i].Path))
}
