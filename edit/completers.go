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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/elves/elvish/edit/lscolors"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
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
}{
	{"index", complIndex},
	{"command", complFormHead},
	{"redir", complRedir},
	{"argument", complArg},
}

// TODO: Replace *eval.Evaler with the smallest possible interface
type completerFinder func(parse.Node, *eval.Evaler) completerIface

var completerFinders = []completerFinder{
	findVariableCompleter,
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

func complIndex(n parse.Node, ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	begin, end, current, q, indexee := findIndexContext(n, ev)

	if begin == -1 {
		return nil, errCompletionUnapplicable
	}

	indexeeValue := purelyEvalPrimary(indexee, ev)
	if indexeeValue == nil {
		return nil, errCannotEvalIndexee
	}
	m, ok := indexeeValue.(eval.IterateKeyer)
	if !ok {
		return nil, errCannotIterateKey
	}

	rawCands := make(chan rawCandidate)
	go func() {
		defer close(rawCands)

		complIndexInner(m, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(ev, matcher,
		current, rawCands, q)
	if err != nil {
		return nil, err
	}
	return &complSpec{begin, end, cands}, nil
}

// Find context information for complIndex. It returns the begin and end for
// compl, the current text of this index and its type, and the indexee node.
//
// Right now we only support cases where there is only one level of indexing,
// e.g. $a[<Tab> is supported but $a[x][<Tab> is not.
func findIndexContext(n parse.Node, ev *eval.Evaler) (int, int, string, parse.PrimaryType, *parse.Primary) {
	if parse.IsSep(n) {
		if parse.IsIndexing(n.Parent()) {
			// We are just after an opening bracket.
			indexing := parse.GetIndexing(n.Parent())
			if len(indexing.Indicies) == 1 {
				return n.End(), n.End(), "", parse.Bareword, indexing.Head
			}
		}
		if parse.IsArray(n.Parent()) {
			array := n.Parent()
			if parse.IsIndexing(array.Parent()) {
				// We are after an existing index and spaces.
				indexing := parse.GetIndexing(array.Parent())
				if len(indexing.Indicies) == 1 {
					return n.End(), n.End(), "", parse.Bareword, indexing.Head
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
						return compound.Begin(), compound.End(), current, primary.Type, indexing.Head
					}
				}
			}
		}
	}

	return -1, -1, "", 0, nil
}

func complIndexInner(m eval.IterateKeyer, rawCands chan rawCandidate) {
	m.IterateKey(func(v eval.Value) bool {
		if keyv, ok := v.(eval.String); ok {
			rawCands <- plainCandidate(keyv)
		}
		return true
	})
}

func complFormHead(n parse.Node, ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	begin, end, head, q := findFormHeadContext(n, ev)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}

	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = complFormHeadInner(head, ev, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(ev, matcher,
		head, rawCands, q)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}
	return &complSpec{begin, end, cands}, nil
}

func findFormHeadContext(n parse.Node, ev *eval.Evaler) (int, int, string, parse.PrimaryType) {
	// Determine if we are starting a new command. There are 3 cases:
	// 1. The whole chunk is empty (nothing entered at all): the leaf is a
	//    Chunk.
	// 2. Just after a newline or semicolon: the leaf is a Sep and its parent is
	//    a Chunk.
	// 3. Just after a pipe: the leaf is a Sep and its parent is a Pipeline.
	if parse.IsChunk(n) {
		return n.End(), n.End(), "", parse.Bareword
	}
	if parse.IsSep(n) {
		parent := n.Parent()
		if parse.IsChunk(parent) || parse.IsPipeline(parent) {
			return n.End(), n.End(), "", parse.Bareword
		}
	}

	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return compound.Begin(), compound.End(), head, primary.Type
				}
			}
		}
	}
	return -1, -1, "", 0
}

func complFormHeadInner(head string, ev *eval.Evaler, rawCands chan<- rawCandidate) error {
	if util.DontSearch(head) {
		return complFilenameInner(head, true, rawCands)
	}

	got := func(s string) {
		rawCands <- plainCandidate(s)
	}
	for special := range eval.IsBuiltinSpecial {
		got(special)
	}
	explode, ns, _ := eval.ParseVariable(head)
	if !explode {
		iterateVariables(ev, ns, func(varname string) {
			if strings.HasPrefix(varname, eval.FnPrefix) {
				got(eval.MakeVariableName(false, ns, varname[len(eval.FnPrefix):]))
			} else {
				got(eval.MakeVariableName(false, ns, varname) + "=")
			}
		})
	}
	ev.EachExternal(func(command string) {
		got(command)
		if strings.HasPrefix(head, "e:") {
			got("e:" + command)
		}
	})
	// TODO Support non-module namespaces.
	for ns := range ev.Global.Uses {
		if head != ns+":" {
			got(ns + ":")
		}
	}
	for ns := range ev.Builtin.Uses {
		if head != ns+":" {
			got(ns + ":")
		}
	}

	return nil
}

// complRedir completes redirection RHS.
func complRedir(n parse.Node, ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	begin, end, current, q := findRedirContext(n, ev)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}

	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = complFilenameInner(current, false, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(ev, matcher,
		current, rawCands, q)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}

	return &complSpec{begin, end, cands}, nil
}

func findRedirContext(n parse.Node, ev *eval.Evaler) (int, int, string, parse.PrimaryType) {
	if parse.IsSep(n) {
		if parse.IsRedir(n.Parent()) {
			return n.End(), n.End(), "", parse.Bareword
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary, ev); compound != nil {
			if parse.IsRedir(compound.Parent()) {
				return compound.Begin(), compound.End(), head, primary.Type
			}
		}
	}
	return -1, -1, "", 0
}

// complArg completes arguments. It identifies the context and then delegates
// the actual completion work to a suitable completer.
func complArg(n parse.Node, ev *eval.Evaler, matcher eval.CallableValue) (*complSpec, error) {
	begin, end, current, q, form := findArgContext(n, ev)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}

	// Find out head of the form and preceding arguments.
	// If Form.Head is not a simple compound, head will be "", just what we want.
	_, head, _ := simpleCompound(form.Head, nil, ev)
	var args []string
	for _, compound := range form.Args {
		if compound.Begin() >= begin {
			break
		}
		ok, arg, _ := simpleCompound(compound, nil, ev)
		if ok {
			// XXX Arguments that are not simple compounds are simply ignored.
			args = append(args, arg)
		}
	}

	words := make([]string, len(args)+2)
	words[0] = head
	words[len(words)-1] = current
	copy(words[1:len(words)-1], args[:])

	rawCands := make(chan rawCandidate)
	collectErr := make(chan error)
	go func() {
		var err error
		defer func() {
			close(rawCands)
			collectErr <- err
		}()

		err = completeArg(words, ev, rawCands)
	}()

	cands, err := ev.Editor.(*Editor).filterAndCookCandidates(ev, matcher,
		current, rawCands, q)
	if ce := <-collectErr; ce != nil {
		return nil, ce
	}
	if err != nil {
		return nil, err
	}

	return &complSpec{begin, end, cands}, nil
}

func findArgContext(n parse.Node, ev *eval.Evaler) (int, int, string, parse.PrimaryType, *parse.Form) {
	if sep, ok := n.(*parse.Sep); ok {
		if form, ok := sep.Parent().(*parse.Form); ok && form.Head != nil {
			return n.End(), n.End(), "", parse.Bareword, form
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary, ev); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head != nil && form.Head != compound {
					return compound.Begin(), compound.End(), head, primary.Type, form
				}
			}
		}
	}
	return -1, -1, "", 0, nil
}

// TODO: getStyle does redundant stats.
func complFilenameInner(head string, executableOnly bool, rawCands chan<- rawCandidate) error {
	dir, fileprefix := filepath.Split(head)
	dirToRead := dir
	if dirToRead == "" {
		dirToRead = "."
	}

	infos, err := ioutil.ReadDir(dirToRead)
	if err != nil {
		return fmt.Errorf("cannot list directory %s: %v", dirToRead, err)
	}

	lsColor := lscolors.GetColorist()
	// Make candidates out of elements that match the file component.
	for _, info := range infos {
		name := info.Name()
		// Show dot files iff file part of pattern starts with dot, and vice
		// versa.
		if dotfile(fileprefix) != dotfile(name) {
			continue
		}
		// Only accept searchable directories and executable files if
		// executableOnly is true.
		if executableOnly && !(info.IsDir() || (info.Mode()&0111) != 0) {
			continue
		}

		// Full filename for source and getStyle.
		full := dir + name

		suffix := " "
		if info.IsDir() {
			suffix = string(filepath.Separator)
		} else if info.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(full)
			if err == nil && stat.IsDir() {
				// Symlink to directory.
				suffix = string(filepath.Separator)
			}
		}

		rawCands <- &complexCandidate{
			stem: full, codeSuffix: suffix,
			style: ui.StylesFromString(lsColor.GetStyle(full)),
		}
	}

	return nil
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
