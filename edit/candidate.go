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
	cook(q parse.PrimaryType) *candidate
}

type plainCandidate string

func (plainCandidate) Kind() string        { return "string" }
func (p plainCandidate) Repr(l int) string { return eval.String(p).Repr(l) }

func (p plainCandidate) cook(q parse.PrimaryType) *candidate {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return &candidate{code: quoted, menu: unstyled(s)}
}

type complexCandidate struct {
	text          string // Used in the code and the menu.
	codeSuffix    string // Appended to the code.
	displaySuffix string // Appended to the display.
	style         styles // Used in the menu.
}

func (c *complexCandidate) Kind() string    { return "map" }
func (c *complexCandidate) Repr(int) string { return "<complex candidate>" }

func (c *complexCandidate) cook(q parse.PrimaryType) *candidate {
	quoted, _ := parse.QuoteAs(c.text, q)
	return &candidate{
		code: quoted + c.codeSuffix,
		menu: styled{c.text + c.displaySuffix, c.style},
	}
}

func cookCandidates(raws []rawCandidate, q parse.PrimaryType) []*candidate {
	cooked := make([]*candidate, len(raws))
	for i, raw := range raws {
		cooked[i] = raw.cook(q)
	}
	return cooked
}
