package edit

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

// Location mode.

var locationFns = map[string]func(*Editor){
	"start": locStart,
}

// PinnedScore is a special value of Score in storedefs.Dir to represent that the
// directory is pinned.
var PinnedScore = math.Inf(1)

type location struct {
	home     string // The home directory; leave empty if unknown.
	all      []storedefs.Dir
	filtered []storedefs.Dir
}

func newLocation(dirs []storedefs.Dir, home string) *listing {
	l := newListing(nil, &location{all: dirs, home: home})
	return &l
}

func (loc *location) ModeTitle(i int) string {
	return " LOCATION "
}

func (*location) CursorOnModeLine() bool {
	return true
}

func (loc *location) Len() int {
	return len(loc.filtered)
}

func (loc *location) Show(i int) (string, ui.Styled) {
	var header string
	score := loc.filtered[i].Score
	if score == PinnedScore {
		header = "*"
	} else {
		header = fmt.Sprintf("%.0f", score)
	}
	return header, ui.Unstyled(showPath(loc.filtered[i].Path, loc.home))
}

func (loc *location) Filter(filter string) int {
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

// Editor interface.

func (loc *location) Accept(i int, ed *Editor) {
	err := eval.Chdir(loc.filtered[i].Path, ed.daemon)
	if err != nil {
		ed.Notify("%v", err)
	}
	ed.SetModeInsert()
}

func locStart(ed *Editor) {
	if ed.daemon == nil {
		ed.Notify("%v", ErrStoreOffline)
		return
	}

	// Pinned directories are also blacklisted to prevent them from showing up
	// twice.
	black := convertListsToSet(ed.locHidden, ed.locPinned)
	pwd, err := os.Getwd()
	if err == nil {
		black[pwd] = struct{}{}
	}
	stored, err := ed.daemon.Dirs(black)
	if err != nil {
		ed.Notify("store error: %v", err)
		return
	}

	// Concatenate pinned and stored dirs, pinned first.
	pinned := convertListToDirs(ed.locPinned)
	dirs := make([]storedefs.Dir, len(pinned)+len(stored))
	copy(dirs, pinned)
	copy(dirs[len(pinned):], stored)

	// Drop the error. When there is an error, home is "", which is used to
	// signify "no home known" in location.
	home, _ := util.GetHome("")
	listing := newLocation(dirs, home)
	listing.binding = &ed.locationBinding
	ed.mode = listing
}

// convertListToDirs converts a list of strings to []storedefs.Dir. It uses the
// special score of PinnedScore to signify that the directory is pinned.
func convertListToDirs(li vector.Vector) []storedefs.Dir {
	pinned := make([]storedefs.Dir, 0, li.Len())
	// XXX(xiaq): silently drops non-string items.
	for it := li.Iterator(); it.HasElem(); it.Next() {
		if s, ok := it.Elem().(string); ok {
			pinned = append(pinned, storedefs.Dir{s, PinnedScore})
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

// Configurations.

type editorLocConfig struct {
	locHidden vector.Vector
	locPinned vector.Vector
}

func init() {
	atEditorInit(func(ed *Editor, ns eval.Ns) {
		ed.locHidden = types.EmptyList
		ns["loc-hidden"] = eval.NewVariableFromPtr(&ed.locHidden)
		ed.locPinned = types.EmptyList
		ns["loc-pinned"] = eval.NewVariableFromPtr(&ed.locPinned)
	})
}
