package complete

import (
	"reflect"

	"github.com/elves/elvish/parse"
)

// Reports whether a and b have the same dynamic type. Useful as a more succinct
// alternative to type assertions.
func is(a, b parse.Node) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}

// Useful as arguments to is.
var (
	aChunk    = &parse.Chunk{}
	aPipeline = &parse.Pipeline{}
	aForm     = &parse.Form{}
	aArray    = &parse.Array{}
	aIndexing = &parse.Indexing{}
	aPrimary  = &parse.Primary{}
	aRedir    = &parse.Redir{}
	aSep      = &parse.Sep{}
)

func primaryInSimpleCompound(pn *parse.Primary, ev PureEvaler) (*parse.Compound, string) {
	indexing, ok := pn.Parent().(*parse.Indexing)
	if !ok {
		return nil, ""
	}
	compound, ok := indexing.Parent().(*parse.Compound)
	if !ok {
		return nil, ""
	}
	head, err := ev.PurelyEvalPartialCompound(compound, indexing)
	if err != nil {
		return nil, ""
	}
	return compound, head
}

func purelyEvalForm(form *parse.Form, seed string, upto int, ev PureEvaler) []string {
	// Find out head of the form and preceding arguments.
	// If form.Head is not a simple compound, head will be "", just what we want.
	head, _ := ev.PurelyEvalPartialCompound(form.Head, nil)
	words := []string{head}
	for _, compound := range form.Args {
		if compound.Range().From >= upto {
			break
		}
		if arg, err := ev.PurelyEvalCompound(compound); err == nil {
			// XXX Arguments that are not simple compounds are simply ignored.
			words = append(words, arg)
		}
	}

	words = append(words, seed)
	return words
}
