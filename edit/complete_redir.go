package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type redirCompleter struct {
	seed       string
	quoting    parse.PrimaryType
	begin, end int
}

func (*redirCompleter) name() string { return "redir" }

func findRedirCompleter(n parse.Node, ev *eval.Evaler) completerIface {
	if parse.IsSep(n) {
		if parse.IsRedir(n.Parent()) {
			return &redirCompleter{"", quotingForEmptySeed, n.End(), n.End()}
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, seed := primaryInSimpleCompound(primary, ev); compound != nil {
			if parse.IsRedir(compound.Parent()) {
				return &redirCompleter{seed, primary.Type, compound.Begin(), compound.End()}
			}
		}
	}
	return nil
}

func (compl *redirCompleter) complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {

	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = complFilenameInner(compl.seed, false, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(
		ev, matcher, compl.seed, rawCands, compl.quoting)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}

	return &complSpec{compl.begin, compl.end, cands}, nil
}
