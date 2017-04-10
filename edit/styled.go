package edit

import (
	"sort"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// styled is a piece of text with style.
type styled struct {
	text   string
	styles styles
}

func unstyled(s string) styled {
	return styled{s, styles{}}
}

func (s *styled) Kind() string {
	return "styled"
}

func (s *styled) String() string {
	return "\033[" + s.styles.String() + "m" + s.text + "\033[m"
}

func (s *styled) Repr(indent int) string {
	return "(le:styled " + parse.Quote(s.text) + " " + parse.Quote(s.styles.String()) + ")"
}

func styledBuiltin(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	var textv, stylev eval.String
	eval.ScanArgs(args, &textv, &stylev)
	text, style := string(textv), string(stylev)
	eval.TakeNoOpt(opts)

	out := ec.OutputChan()
	out <- &styled{text, stylesFromString(style)}
}

// Boilerplates for sorting.

type styleds []styled

func (s styleds) Len() int           { return len(s) }
func (s styleds) Less(i, j int) bool { return s[i].text < s[j].text }
func (s styleds) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func sortStyleds(s []styled) {
	sort.Sort(styleds(s))
}
