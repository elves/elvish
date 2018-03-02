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
	if parse.IsSep(n) {
		if parse.IsIndexing(n.Parent()) {
			// We are just after an opening bracket.
			indexing := parse.GetIndexing(n.Parent())
			if len(indexing.Indicies) == 1 {
				if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
					return &indexComplContext{
						complContextCommon{
							"", quotingForEmptySeed, n.End(), n.End()},
						indexee,
					}
				}
			}
		}
		if parse.IsArray(n.Parent()) {
			array := n.Parent()
			if parse.IsIndexing(array.Parent()) {
				// We are after an existing index and spaces.
				indexing := parse.GetIndexing(array.Parent())
				if len(indexing.Indicies) == 1 {
					if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
						return &indexComplContext{
							complContextCommon{
								"", quotingForEmptySeed, n.End(), n.End()},
							indexee,
						}
					}
				}
			}
		}
	}

	if parse.IsPrimary(n) {
		primary := parse.GetPrimary(n)
		compound, seed := primaryInSimpleCompound(primary, ev)
		if compound != nil {
			if parse.IsArray(compound.Parent()) {
				array := compound.Parent()
				if parse.IsIndexing(array.Parent()) {
					// We are just after an incomplete index.
					indexing := parse.GetIndexing(array.Parent())
					if len(indexing.Indicies) == 1 {
						if indexee := ev.PurelyEvalPrimary(indexing.Head); indexee != nil {
							return &indexComplContext{
								complContextCommon{
									seed, primary.Type, compound.Begin(), compound.End()},
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
