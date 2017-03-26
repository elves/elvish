package edit

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// Location mode.

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
	header := fmt.Sprintf("%.0f", loc.filtered[i].Score)
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
		Logger.Printf("failed to compile regexp %q: %v", b.String(), err)
		return emptyRegexp
	}
	return p
}

// Editor interface.

func (loc *location) Accept(i int, ed *Editor) {
	dir := loc.filtered[i].Path
	err := os.Chdir(dir)
	if err == nil {
		store := ed.store
		go func() {
			store.Waits.Add(1)
			// XXX Error ignored.
			store.AddDir(dir, 1)
			store.Waits.Done()
			Logger.Println("added dir to store:", dir)
		}()
	} else {
		ed.Notify("%v", err)
	}
	ed.mode = &ed.insert
}

func startLocation(ed *Editor) {
	if ed.store == nil {
		ed.Notify("%v", ErrStoreOffline)
		return
	}
	black := convertBlacklist(ed.locationHidden.Get().(eval.List))
	dirs, err := ed.store.GetDirs(black)
	if err != nil {
		ed.Notify("store error: %v", err)
		return
	}

	// Drop the error. When there is an error, home is "", which is used to
	// signify "no home known" in location.
	home, _ := util.GetHome("")
	ed.location = newLocation(dirs, home)
	ed.mode = ed.location
}

func convertBlacklist(li eval.List) map[string]struct{} {
	black := make(map[string]struct{})
	// XXX(xiaq): silently drops non-string items.
	li.Iterate(func(v eval.Value) bool {
		if s, ok := v.(eval.String); ok {
			black[string(s)] = struct{}{}
		}
		return true
	})
	return black
}
