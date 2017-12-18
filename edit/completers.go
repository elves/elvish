package edit

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
// the default setting of prefix matching, it might be easier to define complSpec
// in such a way that completers say "any of aths, id and wd can be appended to
// the 'p' in the source code". However, this is not flexible enough for
// alternative matching mechanism like substring matching or subsequence
// matching, where the "seed" of completion (here, p) may not be a prefix of the
// candidates.
//
// There is one completer that deserves more attention than others, the
// completer for arguments. Unlike other completers, it delegates most of its
// work to argument completers. See the comment in arg_completers.go for
// details.

import (
	"errors"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

var (
	errCompletionUnapplicable = errors.New("completion unapplicable")
	errCannotEvalIndexee      = errors.New("cannot evaluate indexee")
	errCannotIterateKey       = errors.New("indexee does not support iterating keys")
)

// completer takes the current Node (always a leaf in the AST) and an Editor and
// returns a compl. If the completer does not apply to the type of the current
// Node, it should return an error of ErrCompletionUnapplicable.
type completer func(parse.Node, *eval.Evaler, eval.CallableValue) (*complSpec, error)

type completerIface interface {
	name() string
	complete(ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error)
}

// complSpec is the result of a completer, meaning that any of the candidates can
// replace the text in the interval [begin, end).
type complSpec struct {
	begin      int
	end        int
	candidates []*candidate
}

// completers is the list of all completers.
// TODO(xiaq): Make this list programmable.
var completers = []struct {
	name string
	completer
}{}

// TODO: Replace *eval.Evaler with the smallest possible interface
type completerFinder func(parse.Node, *eval.Evaler) completerIface

var completerFinders = []completerFinder{
	findVariableCompleter,
	findCommandCompleter,
	findIndexCompleter,
	findRedirCompleter,
	findArgCompleter,
}

// complete takes a Node and Evaler and tries all completers. It returns the
// name of the completer, and the result and error it gave. If no completer is
// available, it returns an empty completer name.
func complete(n parse.Node, ev *eval.Evaler) (string, *complSpec, error) {
	ed := ev.Editor.(*Editor)
	for _, item := range completers {
		matcher, ok := ed.lookupMatcher(item.name)
		if !ok {
			return item.name, nil, errMatcherMustBeFn
		}

		compl, err := item.completer(n, ev, matcher)
		if compl != nil {
			return item.name, compl, nil
		} else if err != nil && err != errCompletionUnapplicable {
			return item.name, nil, err
		}
	}
	for _, finder := range completerFinders {
		completer := finder(n, ev)
		if completer == nil {
			continue
		}
		name := completer.name()

		matcher, ok := ed.lookupMatcher(name)
		if !ok {
			return name, nil, errMatcherMustBeFn
		}

		compl, err := completer.complete(ev, matcher)
		return name, compl, err
	}
	return "", nil, nil
}
