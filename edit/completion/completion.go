package completion

// Completion in Elvish is organized around the concept of "completers",
// functions that take the current AST Node (the Node that the cursor is at,
// always a leaf in the AST) and an eval.Evaler and returns a specification for
// the completion (a complSpec) -- a list of completion candidates, and which
// part of the source code they can **replace**. When completion is requested,
// the editor calls each completer; it is up to the completer to decide whether
// they apply to the current context. As soon as one completer returns results,
// the remaining completers are not tried.
//
// As an example instance, if the user writes the following and presses Tab:
//
// echo $p
//
// assuming that only the builtin variables $paths, $pid and $pwd are viable
// candidates, one of the completers -- the variable completer -- will return a
// complSpec that means "any of paths, pid and pwd can replace the 'p' in the
// source code".
//
// Note that the "replace" part in the semantics of complSpec is important: in
// the default setting of prefix matching, it might be easier to define
// complSpec in such a way that completers say "any of aths, id and wd can be
// appended to the 'p' in the source code". However, this is not flexible enough
// for alternative matching mechanism like substring matching or subsequence
// matching, where the "seed" of completion (here, p) may not be a prefix of the
// candidates.
//
// There is one completer that deserves more attention than others, the
// completer for arguments. Unlike other completers, it delegates most of its
// work to argument completers. See the comment in arg_completers.go for
// details.

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
)

var logger = util.GetLogger("[edit/completion] ")

type complContext interface {
	name() string
	common() *complContextCommon
	generate(*complEnv, chan<- rawCandidate) error
}

type complContextCommon struct {
	seed       string
	quoting    parse.PrimaryType
	begin, end int
}

func (c *complContextCommon) common() *complContextCommon { return c }

// complEnv contains environment information that may affect candidate
// generation.
type complEnv struct {
	evaler       *eval.Evaler
	matcher      hashmap.Map
	argCompleter hashmap.Map
}

// complSpec is the result of a completion, meaning that any of the candidates
// can replace the text in the interval [begin, end).
type complSpec struct {
	begin      int
	end        int
	candidates []*candidate
}

// A complContextFinder takes the current Node (always a leaf in the AST) and an
// Evaler, and returns a complContext. If the complContext does not apply to the
// type of the current Node, it should return nil.
type complContextFinder func(parse.Node, pureEvaler) complContext

type pureEvaler interface {
	PurelyEvalCompound(*parse.Compound) (string, error)
	PurelyEvalPartialCompound(cn *parse.Compound, upto *parse.Indexing) (string, error)
	PurelyEvalPrimary(*parse.Primary) interface{}
}

var complContextFinders = []complContextFinder{
	findVariableComplContext,
	findCommandComplContext,
	findIndexComplContext,
	findRedirComplContext,
	findArgComplContext,
}

// complete takes a Node and Evaler and tries all complContexts. It returns the
// name of the complContext, and the result and error it gave. If no complContext is
// available, it returns an empty complContext name.
func complete(n parse.Node, env *complEnv) (string, *complSpec, error) {
	for _, finder := range complContextFinders {
		ctx := finder(n, env.evaler)
		if ctx == nil {
			continue
		}
		name := ctx.name()
		ctxCommon := ctx.common()

		matcher, ok := lookupMatcher(env.matcher, name)
		if !ok {
			return name, nil, errMatcherMustBeFn
		}

		chanRawCandidate := make(chan rawCandidate)
		chanErrGenerate := make(chan error)
		go func() {
			err := ctx.generate(env, chanRawCandidate)
			close(chanRawCandidate)
			chanErrGenerate <- err
		}()

		rawCandidates, errFilter := filterRawCandidates(env.evaler, matcher, ctxCommon.seed, chanRawCandidate)
		candidates := make([]*candidate, len(rawCandidates))
		for i, raw := range rawCandidates {
			candidates[i] = raw.cook(ctxCommon.quoting)
		}
		spec := &complSpec{ctxCommon.begin, ctxCommon.end, candidates}
		return name, spec, util.Errors(<-chanErrGenerate, errFilter)

	}
	return "", nil, nil
}
