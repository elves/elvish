// Package location implements the location mode for the editor.
package location

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
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
	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
)

var logger = util.GetLogger("[edit/location] ")

// pinnedScore is a special value of Score in storedefs.Dir to represent that the
// directory is pinned.
var pinnedScore = math.Inf(1)

type mode struct {
	editor     eddefs.Editor
	binding    eddefs.BindingMap
	hidden     vector.Vector
	pinned     vector.Vector
	workspaces hashmap.Map
	matcher    eval.Callable
}

var matchDirPatternBuiltin = eval.NewBuiltinFn("edit:location:match-dir-pattern", matchDirPattern)

// Init initializes the location mode for an Editor.
func Init(ed eddefs.Editor, ns eval.Ns) {
	m := &mode{ed, eddefs.EmptyBindingMap,
		vals.EmptyList, vals.EmptyList, vals.EmptyMap, matchDirPatternBuiltin}

	ns.AddNs("location",
		eval.Ns{
			"binding":    vars.FromPtr(&m.binding),
			"hidden":     vars.FromPtr(&m.hidden),
			"pinned":     vars.FromPtr(&m.pinned),
			"matcher":    vars.FromPtr(&m.matcher),
			"workspaces": vars.FromPtr(&m.workspaces),
		}.AddBuiltinFn("edit:location:", "start", m.start).
			AddFn("match-dir-pattern", matchDirPatternBuiltin))

	ed.Evaler().AddAfterChdir(func(string) {
		store := ed.Daemon()
		if store == nil {
			return
		}
		pwd, err := os.Getwd()
		if err != nil {
			logger.Println("Failed to get pwd in after-chdir hook:", err)
		}
		go addDir(store, pwd, m.workspaces)
	})
}

func addDir(store storedefs.Store, pwd string, workspaces hashmap.Map) {
	err := store.AddDir(pwd, 1)
	if err != nil {
		logger.Println("add dir in after-chdir hook:", err)
		return
	}
	ws := matchWorkspace(pwd, workspaces)
	if ws == nil {
		return
	}
	err = store.AddDir(ws.workspacify(pwd), 1)
	if err != nil {
		logger.Println("add workspacified dir in after-chdir hook:", err)
	}
}

type wsInfo struct {
	name string
	root string
}

func matchWorkspace(dir string, workspaces hashmap.Map) *wsInfo {
	for it := workspaces.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		name, ok := k.(string)
		if !ok {
			// TODO: Surface to user
			logger.Println("$workspaces key not string", k)
			continue
		}
		if strings.HasPrefix(name, "/") {
			// TODO: Surface to user
			logger.Println("$workspaces key starts with /", k)
			continue
		}
		pattern, ok := v.(string)
		if !ok {
			// TODO: Surface to user
			logger.Println("$workspaces value not string", v)
			continue
		}
		if !strings.HasPrefix(pattern, "^") {
			pattern = "^" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			// TODO: Surface to user
			logger.Println("$workspaces pattern invalid", pattern)
			continue
		}
		if ws := re.FindString(dir); ws != "" {
			return &wsInfo{name, ws}
		}
	}
	return nil
}

func (w *wsInfo) workspacify(dir string) string {
	return w.name + dir[len(w.root):]
}

func (w *wsInfo) unworkspacify(dir string) string {
	return w.root + dir[len(w.name):]
}

func (m *mode) start() {
	ed := m.editor

	daemon := ed.Daemon()
	if daemon == nil {
		ed.Notify("store offline, cannot start location mode")
		return
	}

	// Pinned directories are also blacklisted to prevent them from showing up
	// twice.
	black := convertListsToSet(m.hidden, m.pinned)
	pwd, err := os.Getwd()
	if err == nil {
		black[pwd] = struct{}{}
	}
	stored, err := daemon.Dirs(black)
	if err != nil {
		ed.Notify("store error: %v", err)
		return
	}

	// TODO: Move workspace filtering to the daemon.
	ws := matchWorkspace(pwd, m.workspaces)
	wsName := ""
	if ws != nil {
		wsName = ws.name
	}
	wsNameSlash := wsName + string(filepath.Separator)

	var filtered []storedefs.Dir
	for _, dir := range stored {
		if filepath.IsAbs(dir.Path) || (wsName != "" && (dir.Path == wsName || strings.HasPrefix(dir.Path, wsNameSlash))) {
			filtered = append(filtered, dir)
		}
	}

	// Prepend pinned dirs.
	pinnedDirs := convertListToDirs(m.pinned)
	dirs := make([]storedefs.Dir, len(pinnedDirs)+len(filtered))
	copy(dirs, pinnedDirs)
	copy(dirs[len(pinnedDirs):], filtered)

	// Drop the error. When there is an error, home is "", which is used to
	// signify "no home known" in location.
	home, _ := util.GetHome("")
	ed.SetModeListing(m.binding,
		newProvider(dirs, home, ws, ed.Evaler(), m.matcher))
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

type provider struct {
	all      []storedefs.Dir
	filtered []storedefs.Dir
	home     string // The home directory; leave empty if unknown.
	ws       *wsInfo
	ev       *eval.Evaler
	matcher  eval.Callable
}

func newProvider(dirs []storedefs.Dir, home string, ws *wsInfo, ev *eval.Evaler, matcher eval.Callable) *provider {
	return &provider{dirs, nil, home, ws, ev, matcher}
}

func (*provider) ModeTitle(i int) string {
	return " LOCATION "
}

func (*provider) CursorOnModeLine() bool {
	return true
}

func (p *provider) Len() int {
	return len(p.filtered)
}

func (p *provider) Show(i int) (string, ui.Styled) {
	var header string
	score := p.filtered[i].Score
	if score == pinnedScore {
		header = "*"
	} else {
		header = fmt.Sprintf("%.0f", score)
	}
	return header, ui.Unstyled(showPath(p.filtered[i].Path, p.home))
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

func (p *provider) Filter(filter string) int {
	p.filtered = nil

	// TODO: this is just a replica of `filterRawCandidates`.
	matcherInput := make(chan interface{}, len(p.all))
	stopCollector := make(chan struct{})
	go func() {
		defer close(matcherInput)
		for _, item := range p.all {
			select {
			case matcherInput <- showPath(item.Path, p.home):
				logger.Printf("put %s\n", item.Path)
			case <-stopCollector:
				return
			}
		}
	}()
	defer close(stopCollector)

	ports := []*eval.Port{
		{Chan: matcherInput, File: eval.DevNull}, {File: os.Stdout}, {File: os.Stderr}}
	ec := eval.NewTopFrame(p.ev, eval.NewInternalSource("[editor matcher]"), ports)
	args := []interface{}{filter}

	values, err := ec.CaptureOutput(p.matcher, args, eval.NoOpts)
	if err != nil {
		logger.Printf("failed to match %s: %v", filter, err)
		return -1
	} else if got, expect := len(values), len(p.all); got != expect {
		logger.Printf("wrong match count: got %d, want %d", got, expect)
		return -1
	}

	for i, value := range values {
		if vals.Bool(value) {
			p.filtered = append(p.filtered, p.all[i])
		}
	}

	if len(p.filtered) == 0 {
		return -1
	}
	return 0
}

var emptyRegexp = regexp.MustCompile("")

func makeLocationFilterPattern(s string, ignoreCase bool) *regexp.Regexp {
	var b bytes.Buffer
	if ignoreCase {
		b.WriteString("(?i)")
	}
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

func (p *provider) Accept(i int, ed eddefs.Editor) {
	path := p.filtered[i].Path
	if !filepath.IsAbs(path) {
		path = p.ws.unworkspacify(path)
	}
	err := ed.Evaler().Chdir(path)
	if err != nil {
		ed.Notify("%v", err)
	}
	ed.SetModeInsert()
}

func matchDirPattern(fm *eval.Frame, opts eval.RawOptions, pattern string, inputs eval.Inputs) {
	var options struct {
		IgnoreCase bool
	}
	opts.Scan(&options)

	p := makeLocationFilterPattern(pattern, options.IgnoreCase)
	out := fm.OutputChan()
	inputs(func(v interface{}) {
		s, ok := v.(string)
		if !ok {
			logger.Printf("input item must be string, but got %#v", v)
			return
		}
		out <- vals.Bool(p.MatchString(s))
	})
}
