package edit

import (
	"github.com/elves/elvish/eval"
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

func (ctx *redirComplContext) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {

	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = complFilenameInner(ctx.seed, false, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(
		ev, matcher, ctx.seed, rawCands, ctx.quoting)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}

	return &complSpec{ctx.begin, ctx.end, cands}, nil
}
