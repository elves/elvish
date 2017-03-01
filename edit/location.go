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
	home     string
	all      []store.Dir
	filtered []store.Dir
}

func newLocation(dirs []store.Dir) *location {
	loc := &location{all: dirs}
	loc.listing = newListing(modeLocation, loc)
	home, err := util.GetHome("")
	if err == nil {
		loc.home = home
	}
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
	path := loc.filtered[i].Path
	if path == loc.home {
		path = "~"
	} else if strings.HasPrefix(path, loc.home+"/") {
		path = "~/" + parse.Quote(path[len(loc.home)+1:])
	} else {
		path = parse.Quote(path)
	}
	return header, unstyled(path)
}

func (loc *location) Filter(filter string) int {
	loc.filtered = nil
	pattern := makeLocationFilterPattern(filter, loc.home)
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

func makeLocationFilterPattern(s, home string) *regexp.Regexp {
	// First expand tilde.
	if s == "~" {
		s = home
	} else if len(s) >= 2 && s[:2] == "~/" {
		s = home + "/" + s[2:]
	}
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

	ed.location = newLocation(dirs)
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
