package lastcmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse/parseutil"
)

type state struct {
	line     string
	words    []string
	filtered []entry
	minus    bool
}

type entry struct {
	i int
	s string
}

// Init initializes the lastcmd module for an Editor.
func Init(ed eddefs.Editor, ns eval.Ns) {
	lc := &state{}
	binding := eddefs.EmptyBindingMap

	subns := eval.Ns{
		"binding": vars.FromPtr(&binding),
	}
	subns.AddBuiltinFns("edit:lastcmd:", map[string]interface{}{
		"start":       func() { lc.start(ed, binding) },
		"accept-line": func() { lc.acceptLine(ed) },
	})
	ns.AddNs("lastcmd", subns)
}

func newState(line string) *state {
	return &state{line, parseutil.Wordify(line), nil, false}
}

func (*state) AutoAccept() bool {
	return true
}

func (*state) ModeTitle(int) string {
	return " LASTCMD "
}

func (s *state) Len() int {
	return len(s.filtered)
}

func (s *state) Show(i int) (string, ui.Styled) {
	entry := s.filtered[i]
	var head string
	if entry.i == -1 {
		head = "M-1"
	} else if s.minus {
		head = fmt.Sprintf("%d", entry.i-len(s.words))
	} else {
		head = fmt.Sprintf("%d", entry.i)
	}
	return head, ui.Unstyled(entry.s)
}

func (s *state) Filter(filter string) int {
	s.filtered = nil
	s.minus = len(filter) > 0 && filter[0] == '-'
	if filter == "" || filter == "-" {
		s.filtered = append(s.filtered, entry{-1, s.line})
	} else if _, err := strconv.Atoi(filter); err != nil {
		return -1
	}
	// Quite inefficient way to filter by prefix of stringified index.
	n := len(s.words)
	for i, word := range s.words {
		if filter == "" ||
			(!s.minus && strings.HasPrefix(strconv.Itoa(i), filter)) ||
			(s.minus && strings.HasPrefix(strconv.Itoa(i-n), filter)) {
			s.filtered = append(s.filtered, entry{i, word})
		}
	}
	if len(s.filtered) == 0 {
		return -1
	}
	return 0
}

func (s *state) Accept(i int, ed eddefs.Editor) {
	ed.InsertAtDot(s.filtered[i].s)
	ed.SetModeInsert()
}

func (s *state) start(ed eddefs.Editor, binding eddefs.BindingMap) {
	daemon := ed.Daemon()
	if daemon == nil {
		ed.Notify("store offline, cannot start lastcmd mode")
		return
	}
	_, cmd, err := daemon.PrevCmd(-1, "")
	if err != nil {
		ed.Notify("db error: %s", err.Error())
		return
	}
	*s = *newState(cmd)
	ed.SetModeListing(binding, s)
}

func (s *state) acceptLine(ed eddefs.Editor) {
	ed.InsertAtDot(s.line)
	ed.SetModeInsert()
}
