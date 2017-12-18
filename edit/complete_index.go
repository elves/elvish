package edit

import (
	"github.com/elves/elvish/edit/nodeutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type indexComplContext struct {
	seed       string
	quoting    parse.PrimaryType
	indexee    eval.Value
	begin, end int
}

func (*indexComplContext) name() string { return "index" }

// Find context information for complIndex.
//
// Right now we only support cases where there is only one level of indexing,
// e.g. $a[<Tab> is supported but $a[x][<Tab> is not.
func findIndexComplContext(n parse.Node, ev *eval.Evaler) complContext {
	if parse.IsSep(n) {
		if parse.IsIndexing(n.Parent()) {
			// We are just after an opening bracket.
			indexing := parse.GetIndexing(n.Parent())
			if len(indexing.Indicies) == 1 {
				if indexee := nodeutil.PurelyEvalPrimary(indexing.Head, ev); indexee != nil {
					return &indexComplContext{"", quotingForEmptySeed, indexee, n.End(), n.End()}
				}
			}
		}
		if parse.IsArray(n.Parent()) {
			array := n.Parent()
			if parse.IsIndexing(array.Parent()) {
				// We are after an existing index and spaces.
				indexing := parse.GetIndexing(array.Parent())
				if len(indexing.Indicies) == 1 {
					if indexee := nodeutil.PurelyEvalPrimary(indexing.Head, ev); indexee != nil {
						return &indexComplContext{"", quotingForEmptySeed, indexee, n.End(), n.End()}
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
						if indexee := nodeutil.PurelyEvalPrimary(indexing.Head, ev); indexee != nil {
							return &indexComplContext{current, primary.Type, indexee, compound.Begin(), compound.End()}
						}
					}
				}
			}
		}
	}

	return nil
}

func (ctx *indexComplContext) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	m, ok := ctx.indexee.(eval.IterateKeyer)
	if !ok {
		return nil, errCannotIterateKey
	}

	rawCands := make(chan rawCandidate)
	go func() {
		defer close(rawCands)

		complIndexInner(m, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(
		ev, matcher, ctx.seed, rawCands, ctx.quoting)
	if err != nil {
		return nil, err
	}
	return &complSpec{ctx.begin, ctx.end, cands}, nil
}

func complIndexInner(m eval.IterateKeyer, rawCands chan rawCandidate) {
	m.IterateKey(func(v eval.Value) bool {
		if keyv, ok := v.(eval.String); ok {
			rawCands <- plainCandidate(keyv)
		}
		return true
	})
}
