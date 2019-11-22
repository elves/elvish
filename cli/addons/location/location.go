// Package location implements an addon that supports viewing location history
// and changing to a selected directory.
package location

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/combobox"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

// Config is the configuration to start the location history feature.
type Config struct {
	// Binding is the key binding.
	Binding el.Handler
	// Store provides the directory history and the function to change directory.
	Store Store
	// IteratePinned specifies pinned directories by calling the given function
	// with all pinned directories.
	IteratePinned func(func(string))
	// IterateHidden specifies hidden directories by calling the given function
	// with all hidden directories.
	IterateHidden func(func(string))
	// IterateWorksapce specifies workspace configuration.
	IterateWorkspaces WorkspaceIterator
}

// Store defines the interface for interacting with the directory history.
type Store interface {
	Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error)
	Chdir(dir string) error
	Getwd() (string, error)
}

// A special score for pinned directories.
var pinnedScore = math.Inf(1)

// Start starts the directory history feature.
func Start(app cli.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no dir history store")
		return
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
		app.Notify("db error: " + err.Error())
		if len(dirs) == 0 {
			return
		}
	}
	for _, dir := range storedDirs {
		if filepath.IsAbs(dir.Path) {
			dirs = append(dirs, dir)
		} else if wsKind != "" && hasPathPrefix(dir.Path, wsKind) {
			dirs = append(dirs, dir)
		}
	}

	home, _ := util.GetHome("")
	l := list{dirs, home}

	w := combobox.New(combobox.Spec{
		CodeArea: codearea.Spec{
			Prompt: layout.ModePrompt("LOCATION", true),
		},
		ListBox: listbox.Spec{
			OverlayHandler: cfg.Binding,
			OnAccept: func(it listbox.Items, i int) {
				path := it.(list).dirs[i].Path
				if strings.HasPrefix(path, wsKind) {
					path = wsRoot + path[len(wsKind):]
				}
				err := cfg.Store.Chdir(path)
				if err != nil {
					app.Notify(err.Error())
				}
				app.MutateState(func(s *cli.State) { s.Addon = nil })
			},
		},
		OnFilter: func(w combobox.Widget, p string) {
			w.ListBox().Reset(l.filter(p), 0)
		},
	})
	app.MutateState(func(s *cli.State) { s.Addon = w })
	app.Redraw()
}

func hasPathPrefix(path, prefix string) bool {
	return path == prefix ||
		strings.HasPrefix(path, prefix+string(filepath.Separator))
}

// WorkspaceIterator is a function that iterates all workspaces by calling
// the passed function with the name and pattern of each kind of workspace.
// Iteration should stop when the called function returns false.
type WorkspaceIterator func(func(kind, pattern string) bool)

// Parse returns whether the path matches any kind of workspace. If there is
// a match, it returns the kind of the workspace and the root. It there is no
// match, it returns "", "".
func (ws WorkspaceIterator) Parse(path string) (kind, root string) {
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

type list struct {
	dirs []storedefs.Dir
	home string
}

func (l list) filter(p string) list {
	if p == "" {
		return l
	}
	re := makeRegexpForPattern(p)
	var filteredDirs []storedefs.Dir
	for _, dir := range l.dirs {
		if re.MatchString(showPath(dir.Path, l.home)) {
			filteredDirs = append(filteredDirs, dir)
		}
	}
	return list{filteredDirs, l.home}
}

var (
	quotedPathSep = regexp.QuoteMeta(string(os.PathSeparator))
	emptyRe       = regexp.MustCompile("")
)

func makeRegexpForPattern(p string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("(?i).*") // Ignore case, unanchored
	for i, seg := range strings.Split(p, string(os.PathSeparator)) {
		if i > 0 {
			b.WriteString(".*" + quotedPathSep + ".*")
		}
		b.WriteString(regexp.QuoteMeta(seg))
	}
	b.WriteString(".*")
	re, err := regexp.Compile(b.String())
	if err != nil {
		// TODO: Log the error.
		return emptyRe
	}
	return re
}

func (l list) Show(i int) styled.Text {
	return styled.Plain(fmt.Sprintf("%s %s",
		showScore(l.dirs[i].Score), showPath(l.dirs[i].Path, l.home)))
}

func (l list) Len() int { return len(l.dirs) }

func showScore(f float64) string {
	if f == pinnedScore {
		return "  *"
	}
	return fmt.Sprintf("%3.0f", f)
}

func showPath(path, home string) string {
	if path == home {
		return "~"
	} else if strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + path[len(home):]
	}
	return path
}
