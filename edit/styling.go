package edit

import (
	"bytes"
	"sort"

	"github.com/elves/elvish/edit/ui"
)

// Preparing and applying styling.

type styling struct {
	begins []stylingEvent
	ends   []stylingEvent
}

func (s *styling) add(begin, end int, style string) {
	if style == "" {
		return
	}
	s.begins = append(s.begins, stylingEvent{begin, style})
	s.ends = append(s.ends, stylingEvent{end, style})
}

func (s *styling) apply() *stylingApplier {
	sort.Sort(stylingEvents(s.begins))
	sort.Sort(stylingEvents(s.ends))
	return &stylingApplier{s, make(map[string]int), 0, 0, ""}
}

type stylingApplier struct {
	*styling
	occurrence map[string]int
	ibegin     int
	iend       int
	result     string
}

func (a *stylingApplier) at(i int) {
	changed := false
	for a.iend < len(a.ends) && a.ends[a.iend].pos == i {
		a.occurrence[a.ends[a.iend].style]--
		a.iend++
		changed = true
	}
	for a.ibegin < len(a.begins) && a.begins[a.ibegin].pos == i {
		a.occurrence[a.begins[a.ibegin].style]++
		a.ibegin++
		changed = true
	}

	if changed {
		b := new(bytes.Buffer)
		for style, occ := range a.occurrence {
			if occ == 0 {
				continue
			}
			if b.Len() > 0 {
				b.WriteString(";")
			}
			b.WriteString(ui.TranslateStyle(style))
		}
		a.result = b.String()
	}
}

func (a *stylingApplier) get() string {
	return a.result
}

type stylingEvent struct {
	pos   int
	style string
}

type stylingEvents []stylingEvent

func (s stylingEvents) Len() int           { return len(s) }
func (s stylingEvents) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s stylingEvents) Less(i, j int) bool { return s[i].pos < s[j].pos }
