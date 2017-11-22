package edit

import (
	"fmt"
	"os"
	"sort"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

type candidate struct {
	code string    // This is what will be substitued on the command line.
	menu ui.Styled // This is what is displayed in the completion menu.
}

// rawCandidate is what can be converted to a candidate.
type rawCandidate interface {
	eval.Value
	text() string
	cook(q parse.PrimaryType) *candidate
}

type rawCandidates []rawCandidate

func (cs rawCandidates) Len() int           { return len(cs) }
func (cs rawCandidates) Swap(i, j int)      { cs[i], cs[j] = cs[j], cs[i] }
func (cs rawCandidates) Less(i, j int) bool { return cs[i].text() < cs[j].text() }

// plainCandidate is a minimal implementation of rawCandidate.
type plainCandidate eval.String

func (plainCandidate) Kind() string               { return "string" }
func (p plainCandidate) Equal(a interface{}) bool { return p == a }
func (p plainCandidate) Hash() uint32             { return hash.String(string(p)) }
func (p plainCandidate) Repr(l int) string        { return eval.String(p).Repr(l) }
func (p plainCandidate) text() string             { return string(p) }

func (p plainCandidate) cook(q parse.PrimaryType) *candidate {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return &candidate{code: quoted, menu: ui.Unstyled(s)}
}

// noQuoteCandidate is a rawCandidate that does not quote when cooked.
type noQuoteCandidate string

func (noQuoteCandidate) Kind() string                { return "string" }
func (nq noQuoteCandidate) Equal(a interface{}) bool { return nq == a }
func (nq noQuoteCandidate) Hash() uint32             { return hash.String(string(nq)) }
func (nq noQuoteCandidate) Repr(l int) string        { return eval.String(nq).Repr(l) }
func (nq noQuoteCandidate) text() string             { return string(nq) }

func (nq noQuoteCandidate) cook(parse.PrimaryType) *candidate {
	s := string(nq)
	return &candidate{code: s, menu: ui.Unstyled(s)}
}

// complexCandidate is an implementation of rawCandidate that offers
// customization options.
type complexCandidate struct {
	stem          string    // Used in the code and the menu.
	codeSuffix    string    // Appended to the code.
	displaySuffix string    // Appended to the display.
	style         ui.Styles // Used in the menu.
}

func (c *complexCandidate) Kind() string { return "map" }

func (c *complexCandidate) Equal(a interface{}) bool {
	rhs, ok := a.(*complexCandidate)
	return ok && c.stem == rhs.stem && c.codeSuffix == rhs.codeSuffix && c.displaySuffix == rhs.displaySuffix && c.style.Eq(rhs.style)
}

func (c *complexCandidate) Hash() uint32 {
	h := hash.DJBInit
	h = hash.DJBCombine(h, hash.String(c.stem))
	h = hash.DJBCombine(h, hash.String(c.codeSuffix))
	h = hash.DJBCombine(h, hash.String(c.displaySuffix))
	h = hash.DJBCombine(h, c.style.Hash())
	return h
}

func (c *complexCandidate) Repr(indent int) string {
	// TODO(xiaq): Pretty-print when indent >= 0
	return fmt.Sprintf("(edit:complex-candidate %s &code-suffix=%s &display-suffix=%s style=%s)",
		parse.Quote(c.stem), parse.Quote(c.codeSuffix),
		parse.Quote(c.displaySuffix), parse.Quote(c.style.String()))
}

func (c *complexCandidate) text() string { return c.stem }

func (c *complexCandidate) cook(q parse.PrimaryType) *candidate {
	quoted, _ := parse.QuoteAs(c.stem, q)
	return &candidate{
		code: quoted + c.codeSuffix,
		menu: ui.Styled{c.stem + c.displaySuffix, c.style},
	}
}

// outputComplexCandidate composes a complexCandidate.
func outputComplexCandidate(ec *eval.EvalCtx,
	args []eval.Value, opts map[string]eval.Value) {

	var style string
	c := &complexCandidate{}

	eval.ScanArgs(args, &c.stem)
	eval.ScanOpts(opts,
		eval.OptToScan{"code-suffix", &c.codeSuffix, eval.String("")},
		eval.OptToScan{"display-suffix", &c.displaySuffix, eval.String("")},
		eval.OptToScan{"style", &style, eval.String("")},
	)
	if style != "" {
		c.style = ui.StylesFromString(style)
	}

	ec.OutputChan() <- c
}

func (ed *Editor) filterAndCookCandidates(ev *eval.Evaler, matcher eval.CallableValue,
	pattern string, cands chan rawCandidate, q parse.PrimaryType) ([]*candidate, error) {

	input := make(chan eval.Value)
	var rcs []rawCandidate
	go func() {
		defer close(input)
		for rc := range cands {
			rcs = append(rcs, rc)
			input <- eval.String(rc.text())
		}
	}()
	// drain up input channel, make sure goroutine exits actually
	defer func() {
		for range input {
		}
	}()

	ports := []*eval.Port{
		{Chan: input, File: eval.DevNull}, {File: os.Stdout}, {File: os.Stderr}}
	ec := eval.NewTopEvalCtx(ev, "[editor matcher]", "", ports)

	args := []eval.Value{eval.String(pattern)}
	values, err := ec.PCaptureOutput(matcher, args, eval.NoOpts)
	if err != nil {
		return nil, err
	} else if len(values) != len(rcs) {
		return nil, errIncorrectNumOfResults
	}

	var (
		filtered []*candidate
		frcs     rawCandidates
	)
	for i, value := range values {
		if eval.ToBool(value) {
			frcs = append(frcs, rcs[i])
		}
	}
	sort.Sort(frcs)
	for _, rc := range frcs {
		filtered = append(filtered, rc.cook(q))
	}

	return filtered, nil
}
