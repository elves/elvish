package edit

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// Location mode.

var _ = registerListingBuiltins("loc", map[string]func(*Editor){
	"start": locStart,
}, func(ed *Editor) *listing { return &ed.location.listing })

func init() {
	registerListingBindings(modeLocation, "loc", map[ui.Key]string{})
}

// PinnedScore is a special value of Score in store.Dir to represent that the
// directory is pinned.
var PinnedScore = math.Inf(1)

type location struct {
	listing
	home     string // The home directory; leave empty if unknown.
	all      []store.Dir
	filtered []store.Dir
}

func newLocation(dirs []store.Dir, home string) *location {
	loc := &location{all: dirs, home: home}
	loc.listing = newListing(modeLocation, loc)
	return loc
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

func (loc *location) Show(i int) (string, styled) {
	var header string
	score := loc.filtered[i].Score
	if score == PinnedScore {
		header = "*"
	} else {
		header = fmt.Sprintf("%.0f", score)
	}
	return header, unstyled(showPath(loc.filtered[i].Path, loc.home))
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

func (ed *Editor) chdir(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	err = os.Chdir(dir)
	if err == nil {
		store := ed.store
		store.Waits.Add(1)
		go func() {
			// XXX Error ignored.
			store.AddDir(dir, 1)
			store.Waits.Done()
			logger.Println("added dir to store:", dir)
		}()
	}
	return err
}

// Editor interface.

func (loc *location) Accept(i int, ed *Editor) {
	err := ed.chdir(loc.filtered[i].Path)
	if err != nil {
		ed.Notify("%v", err)
	}
	ed.mode = &ed.insert
}

func locStart(ed *Editor) {
	if ed.store == nil {
		ed.Notify("%v", ErrStoreOffline)
		return
	}
	black := convertListToSet(ed.locationHidden.Get().(eval.List))
	dirs, err := ed.store.GetDirs(black)
	if err != nil {
		ed.Notify("store error: %v", err)
		return
	}

	pinnedValue := ed.locationPinned.Get().(eval.List)
	pinned := convertListToDirs(pinnedValue)
	pinnedSet := convertListToSet(pinnedValue)

	// TODO(xiaq): Optimize this by changing GetDirs to a callback API, and
	// build dirs by first putting pinned directories and then appending those
	// from store.
	for _, d := range dirs {
		_, inPinned := pinnedSet[d.Path]
		if !inPinned {
			pinned = append(pinned, d)
		}
	}
	dirs = pinned

	// Drop the error. When there is an error, home is "", which is used to
	// signify "no home known" in location.
	home, _ := util.GetHome("")
	ed.location = newLocation(dirs, home)
	ed.mode = ed.location
}

// convertListToDirs converts a list of strings to []store.Dir. It uses the
// special score of PinnedScore to signify that the directory is pinned.
func convertListToDirs(li eval.List) []store.Dir {
	pinned := make([]store.Dir, 0, li.Len())
	// XXX(xiaq): silently drops non-string items.
	li.Iterate(func(v eval.Value) bool {
		if s, ok := v.(eval.String); ok {
			pinned = append(pinned, store.Dir{string(s), PinnedScore})
		}
		return true
	})
	return pinned
}

func convertListToSet(li eval.List) map[string]struct{} {
	set := make(map[string]struct{})
	// XXX(xiaq): silently drops non-string items.
	li.Iterate(func(v eval.Value) bool {
		if s, ok := v.(eval.String); ok {
			set[string(s)] = struct{}{}
		}
		return true
	})
	return set
}
