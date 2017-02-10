package edit

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
)

// Location mode.

type location struct {
	listing
	all      []store.Dir
	filtered []store.Dir
}

func newLocation(dirs []store.Dir) *location {
	loc := &location{all: dirs}
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
	cand := loc.filtered[i]
	return fmt.Sprintf("%.0f", cand.Score), unstyled(parse.Quote(cand.Path))
}

func (loc *location) Filter(filter string) int {
	loc.filtered = nil
	pattern := makeLocationFilterPattern(filter)
	for _, item := range loc.all {
		if pattern.MatchString(item.Path) {
			loc.filtered = append(loc.filtered, item)
		}
	}

	if len(loc.filtered) == 0 {
		return -1
	}
	return 0
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
	dirs, err := ed.store.ListDirs()
	if err != nil {
		ed.Notify("store error: %v", err)
		return
	}

	ed.location = newLocation(dirs)
	ed.mode = ed.location
}
