package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type candidate struct {
	code string // This is what will be substitued on the command line.
	menu styled // This is what is displayed in the completion menu.
}

// rawCandidate is what can be converted to a candidate.
type rawCandidate interface {
	eval.Value
	text() string
	cook(q parse.PrimaryType) *candidate
}

type plainCandidate string

func (plainCandidate) Kind() string        { return "string" }
func (p plainCandidate) Repr(l int) string { return eval.String(p).Repr(l) }

func (p plainCandidate) text() string { return string(p) }

func (p plainCandidate) cook(q parse.PrimaryType) *candidate {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return &candidate{code: quoted, menu: unstyled(s)}
}

type complexCandidate struct {
	stem          string // Used in the code and the menu.
	codeSuffix    string // Appended to the code.
	displaySuffix string // Appended to the display.
	style         styles // Used in the menu.
}

func (c *complexCandidate) Kind() string    { return "map" }
func (c *complexCandidate) Repr(int) string { return "<complex candidate>" }

func (c *complexCandidate) text() string { return c.stem }

func (c *complexCandidate) cook(q parse.PrimaryType) *candidate {
	quoted, _ := parse.QuoteAs(c.stem, q)
	return &candidate{
		code: quoted + c.codeSuffix,
		menu: styled{c.stem + c.displaySuffix, c.style},
	}
}

func cookCandidates(raws []rawCandidate, pattern string,
	match func(string, string) bool, q parse.PrimaryType) []*candidate {

	var cooked []*candidate
	for _, raw := range raws {
		if match(raw.text(), pattern) {
			cooked = append(cooked, raw.cook(q))
		}
	}
	return cooked
}
