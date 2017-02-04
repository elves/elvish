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

func (loc *location) Len() int {
	return len(loc.filtered)
}

func (loc *location) Show(i int) styled {
	cand := loc.filtered[i]
	return unstyled(fmt.Sprintf("%4.0f %s", cand.Score, parse.Quote(cand.Path)))
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

func (loc *location) ModeTitle(i int) string {
	return " LOCATION "
}

func startLocation(ed *Editor) {
	loc, err := newLocation(ed.store)
	if err != nil {
		ed.Notify("%v", err)
		return
	}

	ed.location = loc
	ed.mode = ed.location
}

func newLocation(s *store.Store) (*location, error) {
	if s == nil {
		return nil, ErrStoreOffline
	}
	all, err := s.ListDirs()
	if err != nil {
		return nil, err
	}

	loc := &location{all: all}
	loc.listing = newListing(modeLocation, loc)
	return loc, nil
}
