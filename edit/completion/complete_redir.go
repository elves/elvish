package completion

import (
	"github.com/elves/elvish/parse"
)

type redirComplContext struct {
	complContextCommon
}

func (*redirComplContext) name() string { return "redir" }

func findRedirComplContext(n parse.Node, ev pureEvaler) complContext {
	if isSep(n) {
		if is(n.Parent(), &parse.Redir{}) {
			return &redirComplContext{complContextCommon{
				"", quotingForEmptySeed, n.Range().To, n.Range().To}}
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if is(compound.Parent(), &parse.Redir{}) {
				return &redirComplContext{complContextCommon{
					seed, primary.Type, compound.Range().From, compound.Range().To}}
			}
		}
	}
	return nil
}

func (ctx *redirComplContext) generate(env *complEnv, ch chan<- rawCandidate) error {
	return complFilenameInner(ctx.seed, false, ch)
}
