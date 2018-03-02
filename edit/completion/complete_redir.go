package completion

import (
	"github.com/elves/elvish/parse"
)

type redirComplContext struct {
	complContextCommon
}

func (*redirComplContext) name() string { return "redir" }

func findRedirComplContext(n parse.Node, ev pureEvaler) complContext {
	if parse.IsSep(n) {
		if parse.IsRedir(n.Parent()) {
			return &redirComplContext{complContextCommon{
				"", quotingForEmptySeed, n.End(), n.End()}}
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if parse.IsRedir(compound.Parent()) {
				return &redirComplContext{complContextCommon{
					seed, primary.Type, compound.Begin(), compound.End()}}
			}
		}
	}
	return nil
}

func (ctx *redirComplContext) generate(env *complEnv, ch chan<- rawCandidate) error {
	return complFilenameInner(ctx.seed, false, ch)
}
