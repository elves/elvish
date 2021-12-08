package modes

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

// Location is a mode for viewing location history and changing to a selected
// directory. It is based on the ComboBox widget.
type Location interface {
	tk.ComboBox
}

// LocationSpec is the configuration to start the location history feature.
type LocationSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// Store provides the directory history and the function to change directory.
	Store LocationStore
	// IteratePinned specifies pinned directories by calling the given function
	// with all pinned directories.
	IteratePinned func(func(string))
	// IterateHidden specifies hidden directories by calling the given function
	// with all hidden directories.
	IterateHidden func(func(string))
	// IterateWorksapce specifies workspace configuration.
	IterateWorkspaces LocationWSIterator
	// Configuration for the filter.
	Filter FilterSpec
}

// LocationStore defines the interface for interacting with the directory history.
type LocationStore interface {
	Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error)
	Chdir(dir string) error
	Getwd() (string, error)
}

// A special score for pinned directories.
var pinnedScore = math.Inf(1)

var errNoDirectoryHistoryStore = errors.New("no directory history store")

// NewLocation creates a new location mode.
func NewLocation(app cli.App, cfg LocationSpec) (Location, error) {
	if cfg.Store == nil {
		return nil, errNoDirectoryHistoryStore
	}

	dirs := []storedefs.Dir{}
	blacklist := map[string]struct{}{}
	wsKind, wsRoot := "", ""

	if cfg.IteratePinned != nil {
		cfg.IteratePinned(func(s string) {
			blacklist[s] = struct{}{}
			dirs = append(dirs, storedefs.Dir{Score: pinnedScore, Path: s})
		})
	}
	if cfg.IterateHidden != nil {
		cfg.IterateHidden(func(s string) { blacklist[s] = struct{}{} })
	}
	wd, err := cfg.Store.Getwd()
	if err == nil {
		blacklist[wd] = struct{}{}
		if cfg.IterateWorkspaces != nil {
			wsKind, wsRoot = cfg.IterateWorkspaces.Parse(wd)
		}
	}
	storedDirs, err := cfg.Store.Dirs(blacklist)
	if err != nil {
		return nil, fmt.Errorf("db error: %v", err)
	}
	for _, dir := range storedDirs {
		if filepath.IsAbs(dir.Path) {
			dirs = append(dirs, dir)
		} else if wsKind != "" && hasPathPrefix(dir.Path, wsKind) {
			dirs = append(dirs, dir)
		}
	}

	l := locationList{dirs}

	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{
			Prompt:      modePrompt(" LOCATION ", true),
			Highlighter: cfg.Filter.Highlighter,
		},
		ListBox: tk.ListBoxSpec{
			Bindings: cfg.Bindings,
			OnAccept: func(it tk.Items, i int) {
				path := it.(locationList).dirs[i].Path
				if strings.HasPrefix(path, wsKind) {
					path = wsRoot + path[len(wsKind):]
				}
				err := cfg.Store.Chdir(path)
				if err != nil {
					app.Notify(ErrorText(err))
				}
				app.PopAddon()
			},
		},
		OnFilter: func(w tk.ComboBox, p string) {
			w.ListBox().Reset(l.filter(cfg.Filter.makePredicate(p)), 0)
		},
	})
	return w, nil
}

func hasPathPrefix(path, prefix string) bool {
	return path == prefix ||
		strings.HasPrefix(path, prefix+string(filepath.Separator))
}

// LocationWSIterator is a function that iterates all workspaces by calling
// the passed function with the name and pattern of each kind of workspace.
// Iteration should stop when the called function returns false.
type LocationWSIterator func(func(kind, pattern string) bool)

// Parse returns whether the path matches any kind of workspace. If there is
// a match, it returns the kind of the workspace and the root. It there is no
// match, it returns "", "".
func (ws LocationWSIterator) Parse(path string) (kind, root string) {
	var foundKind, foundRoot string
	ws(func(kind, pattern string) bool {
		if !strings.HasPrefix(pattern, "^") {
			pattern = "^" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			// TODO(xiaq): Surface the error.
			return true
		}
		if root := re.FindString(path); root != "" {
			foundKind, foundRoot = kind, root
			return false
		}
		return true
	})
	return foundKind, foundRoot
}

type locationList struct {
	dirs []storedefs.Dir
}

func (l locationList) filter(p func(string) bool) locationList {
	var filteredDirs []storedefs.Dir
	for _, dir := range l.dirs {
		if p(fsutil.TildeAbbr(dir.Path)) {
			filteredDirs = append(filteredDirs, dir)
		}
	}
	return locationList{filteredDirs}
}

func (l locationList) Show(i int) ui.Text {
	return ui.T(fmt.Sprintf("%s %s",
		showScore(l.dirs[i].Score), fsutil.TildeAbbr(l.dirs[i].Path)))
}

func (l locationList) Len() int { return len(l.dirs) }

func showScore(f float64) string {
	if f == pinnedScore {
		return "  *"
	}
	return fmt.Sprintf("%3.0f", f)
}
