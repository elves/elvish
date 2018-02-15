// Package location implements the location mode for the editor.
package location

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

var logger = util.GetLogger("[edit/location] ")

// pinnedScore is a special value of Score in storedefs.Dir to represent that the
// directory is pinned.
var pinnedScore = math.Inf(1)

type config struct {
	binding eddefs.BindingMap
	hidden  vector.Vector
	pinned  vector.Vector
}

type state struct {
	home     string // The home directory; leave empty if unknown.
	all      []storedefs.Dir
	filtered []storedefs.Dir
}

// Init initializes the location mode for an Editor.
func Init(ed eddefs.Editor, ns eval.Ns) {
	cfg := &config{
		binding: eddefs.EmptyBindingMap,
		hidden:  vals.EmptyList,
		pinned:  vals.EmptyList,
	}

	subns := eval.Ns{
		"binding": vars.NewFromPtr(&cfg.binding),
		"hidden":  vars.NewFromPtr(&cfg.hidden),
		"pinned":  vars.NewFromPtr(&cfg.pinned),
	}
	subns.AddBuiltinFns("edit:location:", map[string]interface{}{
		"start": func() { locStart(ed, cfg.hidden, cfg.pinned, cfg.binding) },
	})
	ns.AddNs("location", subns)
}

func newState(dirs []storedefs.Dir, home string) *state {
	return &state{all: dirs, home: home}
}

func (loc *state) ModeTitle(i int) string {
	return " LOCATION "
}

func (*state) CursorOnModeLine() bool {
	return true
}

func (loc *state) Len() int {
	return len(loc.filtered)
}

func (loc *state) Show(i int) (string, ui.Styled) {
	var header string
	score := loc.filtered[i].Score
	if score == pinnedScore {
		header = "*"
	} else {
		header = fmt.Sprintf("%.0f", score)
	}
	return header, ui.Unstyled(showPath(loc.filtered[i].Path, loc.home))
}

func (loc *state) Filter(filter string) int {
	loc.filtered = nil
	pattern := makeLocationFilterPattern(filter)
	for _, item := range loc.all {
		if pattern.MatchString(showPath(item.Path, loc.home)) {
			loc.filtered = append(loc.filtered, item)
		}
	}

	if len(loc.filtered) == 0 {
		return -1
	}
	return 0
}

func showPath(path, home string) string {
	if home != "" && path == home {
		return "~"
	} else if home != "" && strings.HasPrefix(path, home+"/") {
		return "~/" + parse.Quote(path[len(home)+1:])
	} else {
		return parse.Quote(path)
	}
}

var emptyRegexp = regexp.MustCompile("")

func makeLocationFilterPattern(s string) *regexp.Regexp {
	var b bytes.Buffer
	b.WriteString(".*")
	segs := strings.Split(s, "/")
	for i, seg := range segs {
		if i > 0 {
			b.WriteString(".*/.*")
		}
		b.WriteString(regexp.QuoteMeta(seg))
	}
	b.WriteString(".*")
	p, err := regexp.Compile(b.String())
	if err != nil {
		logger.Printf("failed to compile regexp %q: %v", b.String(), err)
		return emptyRegexp
	}
	return p
}

func (loc *state) Accept(i int, ed eddefs.Editor) {
	err := eval.Chdir(loc.filtered[i].Path, ed.Daemon())
	if err != nil {
		ed.Notify("%v", err)
	}
	ed.SetModeInsert()
}

func locStart(ed eddefs.Editor, hidden, pinned vector.Vector, binding eddefs.BindingMap) {
	daemon := ed.Daemon()
	if daemon == nil {
		ed.Notify("store offline, cannot start location mode")
		return
	}

	// Pinned directories are also blacklisted to prevent them from showing up
	// twice.
	black := convertListsToSet(hidden, pinned)
	pwd, err := os.Getwd()
	if err == nil {
		black[pwd] = struct{}{}
	}
	stored, err := daemon.Dirs(black)
	if err != nil {
		ed.Notify("store error: %v", err)
		return
	}

	// Concatenate pinned and stored dirs, pinned first.
	pinnedDirs := convertListToDirs(pinned)
	dirs := make([]storedefs.Dir, len(pinnedDirs)+len(stored))
	copy(dirs, pinnedDirs)
	copy(dirs[len(pinnedDirs):], stored)

	// Drop the error. When there is an error, home is "", which is used to
	// signify "no home known" in location.
	home, _ := util.GetHome("")
	ed.SetModeListing(binding, newState(dirs, home))
}

// convertListToDirs converts a list of strings to []storedefs.Dir. It uses the
// special score of pinnedScore to signify that the directory is pinned.
func convertListToDirs(li vector.Vector) []storedefs.Dir {
	pinned := make([]storedefs.Dir, 0, li.Len())
	// XXX(xiaq): silently drops non-string items.
	for it := li.Iterator(); it.HasElem(); it.Next() {
		if s, ok := it.Elem().(string); ok {
			pinned = append(pinned, storedefs.Dir{s, pinnedScore})
		}
	}
	return pinned
}

func convertListsToSet(lis ...vector.Vector) map[string]struct{} {
	set := make(map[string]struct{})
	// XXX(xiaq): silently drops non-string items.
	for _, li := range lis {
		for it := li.Iterator(); it.HasElem(); it.Next() {
			if s, ok := it.Elem().(string); ok {
				set[s] = struct{}{}
			}
		}
	}
	return set
}
