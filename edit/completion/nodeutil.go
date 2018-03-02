package completion

import (
	"github.com/elves/elvish/parse"
)

// Utilities for insepcting the AST. Used for completers and stylists.

func primaryInSimpleCompound(pn *parse.Primary, ev pureEvaler) (*parse.Compound, string) {
	indexing := parse.GetIndexing(pn.Parent())
	if indexing == nil {
		return nil, ""
	}
	compound := parse.GetCompound(indexing.Parent())
	if compound == nil {
		return nil, ""
	}
	head, err := ev.PurelyEvalPartialCompound(compound, indexing)
	if err != nil {
		return nil, ""
	}
	return compound, head
}
