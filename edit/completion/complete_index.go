package completion

import (
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
)

type indexComplContext struct {
	complContextCommon
	indexee interface{}
}

func (*indexComplContext) name() string { return "index" }

// Find context information for complIndex.
//
// Right now we only support cases where there is only one level of indexing,
// e.g. $a[<Tab> is supported but $a[x][<Tab> is not.
func findIndexComplContext(n parse.Node, ev pureEvaler) complContext {
	if is(n, aSep) {
		if is(n.Parent(), aIndexing) {
			// We are just after an opening bracket.
			indexing := n.Parent().(*parse.Indexing)
			if len(indexing.Indicies) == 1 {
				if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
					return &indexComplContext{
						complContextCommon{
							"", quotingForEmptySeed, n.Range().To, n.Range().To},
						indexee,
					}
				}
			}
		}
		if is(n.Parent(), aArray) {
			array := n.Parent()
			if is(array.Parent(), aIndexing) {
				// We are after an existing index and spaces.
				indexing := array.Parent().(*parse.Indexing)
				if len(indexing.Indicies) == 1 {
					if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
						return &indexComplContext{
							complContextCommon{
								"", quotingForEmptySeed, n.Range().To, n.Range().To},
							indexee,
						}
					}
				}
			}
		}
	}

	if is(n, aPrimary) {
		primary := n.(*parse.Primary)
		compound, seed := primaryInSimpleCompound(primary, ev)
		if compound != nil {
			if is(compound.Parent(), aArray) {
				array := compound.Parent()
				if is(array.Parent(), aIndexing) {
					// We are just after an incomplete index.
					indexing := array.Parent().(*parse.Indexing)
					if len(indexing.Indicies) == 1 {
						if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
							return &indexComplContext{
								complContextCommon{
									seed, primary.Type, compound.Range().From, compound.Range().To},
								indexee,
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (ctx *indexComplContext) generate(env *complEnv, ch chan<- rawCandidate) error {
	return vals.IterateKeys(ctx.indexee, func(k interface{}) bool {
		if kstring, ok := k.(string); ok {
			ch <- plainCandidate(kstring)
		}
		return true
	})
}
