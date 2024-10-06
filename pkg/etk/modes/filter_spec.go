package modes

import (
	"strings"

	"src.elv.sh/pkg/ui"
)

// FilterSpec specifies the configuration for the filter in listing modes.
type FilterSpec struct {
	// Called with the filter text to get the filter predicate. If nil, the
	// predicate performs substring match.
	Maker func(string) func(string) bool
	// Highlighter for the filter. If nil, the filter will not be highlighted.
	Highlighter func(string) (ui.Text, []ui.Text)
}

func (f FilterSpec) makePredicate(p string) func(string) bool {
	if f.Maker == nil {
		return func(s string) bool { return strings.Contains(s, p) }
	}
	return f.Maker(p)
}
