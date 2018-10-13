package completion

import (
	"github.com/elves/elvish/parse"
)

// Utilities for insepcting the AST. Used for completers and stylists.

func primaryInSimpleCompound(pn *parse.Primary, ev pureEvaler) (*parse.Compound, string) {
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
