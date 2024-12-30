package edit

import (
	"fmt"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

func startLocation(ed *Editor, c etk.Context) {
	dirs, err := ed.store.Dirs(map[string]struct{}{})
	if err != nil {
		// TODO: report
		return
	}

	pushAddon(c, etk.WithInit(comps.ComboBox,
		"query/prompt", addonPromptText(" LOCATION "),
		"gen-list", func(f string) (comps.ListItems, int) {
			return locationItems{dirs}, 0
		},
		"binding", etkBindingFromBindingMap(ed, &ed.locationBinding),
	), 1)
}

type locationItems struct {
	dirs []storedefs.Dir
}

func (l locationItems) filter(p func(string) bool) locationItems {
	var filteredDirs []storedefs.Dir
	for _, dir := range l.dirs {
		if p(fsutil.TildeAbbr(dir.Path)) {
			filteredDirs = append(filteredDirs, dir)
		}
	}
	return locationItems{filteredDirs}
}

func (l locationItems) Len() int      { return len(l.dirs) }
func (l locationItems) Get(i int) any { return l.dirs[i] }

func (l locationItems) Show(i int) ui.Text {
	return ui.T(fmt.Sprintf("%s %s",
		showScore(l.dirs[i].Score), fsutil.TildeAbbr(l.dirs[i].Path)))
}

func showScore(f float64) string {
	/*
		if f == pinnedScore {
			return "  *"
		}
	*/
	return fmt.Sprintf("%3.0f", f)
}
