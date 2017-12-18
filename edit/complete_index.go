package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type indexCompleter struct {
	seed       string
	quoting    parse.PrimaryType
	indexee    eval.Value
	begin, end int
}

func (*indexCompleter) name() string { return "index" }

// Find context information for complIndex.
//
// Right now we only support cases where there is only one level of indexing,
// e.g. $a[<Tab> is supported but $a[x][<Tab> is not.
func findIndexCompleter(n parse.Node, ev *eval.Evaler) completerIface {
	if parse.IsSep(n) {
		if parse.IsIndexing(n.Parent()) {
			// We are just after an opening bracket.
			indexing := parse.GetIndexing(n.Parent())
			if len(indexing.Indicies) == 1 {
				if indexee := purelyEvalPrimary(indexing.Head, ev); indexee != nil {
					return &indexCompleter{"", quotingForEmptySeed, indexee, n.End(), n.End()}
				}
			}
		}
		if parse.IsArray(n.Parent()) {
			array := n.Parent()
			if parse.IsIndexing(array.Parent()) {
				// We are after an existing index and spaces.
				indexing := parse.GetIndexing(array.Parent())
				if len(indexing.Indicies) == 1 {
					if indexee := purelyEvalPrimary(indexing.Head, ev); indexee != nil {
						return &indexCompleter{"", quotingForEmptySeed, indexee, n.End(), n.End()}
					}
				}
			}
		}
	}

	if parse.IsPrimary(n) {
		primary := parse.GetPrimary(n)
		compound, current := primaryInSimpleCompound(primary, ev)
		if compound != nil {
			if parse.IsArray(compound.Parent()) {
				array := compound.Parent()
				if parse.IsIndexing(array.Parent()) {
					// We are just after an incomplete index.
					indexing := parse.GetIndexing(array.Parent())
					if len(indexing.Indicies) == 1 {
						if indexee := purelyEvalPrimary(indexing.Head, ev); indexee != nil {
							return &indexCompleter{current, primary.Type, indexee, compound.Begin(), compound.End()}
						}
					}
				}
			}
		}
	}

	return nil
}

func (compl *indexCompleter) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	m, ok := compl.indexee.(eval.IterateKeyer)
	if !ok {
		return nil, errCannotIterateKey
	}

	rawCands := make(chan rawCandidate)
	go func() {
		defer close(rawCands)

		complIndexInner(m, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(
		ev, matcher, compl.seed, rawCands, compl.quoting)
	if err != nil {
		return nil, err
	}
	return &complSpec{compl.begin, compl.end, cands}, nil
}

func complIndexInner(m eval.IterateKeyer, rawCands chan rawCandidate) {
	m.IterateKey(func(v eval.Value) bool {
		if keyv, ok := v.(eval.String); ok {
			rawCands <- plainCandidate(keyv)
		}
		return true
	})
}
