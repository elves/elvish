package edit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

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
type completer func(parse.Node, *eval.Evaler) (*compl, error)

// compl is the result of a completer, meaning that any of the candidates can
// replace the text in the interval [begin, end).
type compl struct {
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
	{"variable", complVariable},
	{"index", complIndex},
	{"command name", complFormHead},
	{"redir", complRedir},
	{"argument", complArg},
}

// complete takes a Node and Evaler and tries all completers. It returns the
// name of the completer, and the result and error it gave. If no completer is
// available, it returns an empty completer name.
func complete(n parse.Node, ev *eval.Evaler) (string, *compl, error) {
	for _, item := range completers {
		compl, err := item.completer(n, ev)
		if compl != nil {
			return item.name, compl, nil
		} else if err != nil && err != errCompletionUnapplicable {
			return item.name, nil, err
		}
	}
	return "", nil, nil
}

// TODO(xiaq): Rewrite this to use cookCandidates
func complVariable(n parse.Node, ev *eval.Evaler) (*compl, error) {
	primary := parse.GetPrimary(n)
	if primary == nil || primary.Type != parse.Variable {
		return nil, errCompletionUnapplicable
	}

	// XXX Repeats eval.ParseVariable.
	explode, qname := eval.ParseVariableSplice(primary.Value)
	nsPart, nameHead := eval.ParseVariableQName(qname)
	ns := nsPart
	if len(ns) > 0 {
		ns = ns[:len(ns)-1]
	}

	// Collect matching variables.
	var varnames []string
	iterateVariables(ev, ns, func(varname string) {
		if strings.HasPrefix(varname, nameHead) {
			varnames = append(varnames, nsPart+varname)
		}
	})
	// Collect namespace prefixes.
	// TODO Support non-module namespaces.
	for ns := range ev.Modules {
		if hasProperPrefix(ns+":", qname) {
			varnames = append(varnames, ns+":")
		}
	}
	sort.Strings(varnames)

	cands := make([]*candidate, len(varnames))
	// Build candidates.
	for i, varname := range varnames {
		text := "$" + explode + varname
		cands[i] = &candidate{code: text, menu: unstyled(text)}
	}

	return &compl{n.Begin(), n.End(), cands}, nil
}

func hasProperPrefix(s, p string) bool {
	return len(s) > len(p) && strings.HasPrefix(s, p)
}

func iterateVariables(ev *eval.Evaler, ns string, f func(string)) {
	switch ns {
	case "":
		for varname := range eval.Builtin() {
			f(varname)
		}
		for varname := range ev.Global {
			f(varname)
		}
		// TODO Include local names as well.
	case "E":
		for _, s := range os.Environ() {
			f(s[:strings.IndexByte(s, '=')])
		}
	default:
		// TODO Support non-module namespaces.
		for varname := range ev.Modules[ns] {
			f(varname)
		}
	}
}

func complIndex(n parse.Node, ev *eval.Evaler) (*compl, error) {
	begin, end, current, q, indexee := findIndexContext(n)

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

	cands := complIndexInner(m)

	return &compl{begin, end, cookCandidates(cands, current, q)}, nil
}

// Find context information for complIndex. It returns the begin and end for
// compl, the current text of this index and its type, and the indexee node.
//
// Right now we only support cases where there is only one level of indexing,
// e.g. $a[<Tab> is supported but $a[x][<Tab> is not.
func findIndexContext(n parse.Node) (int, int, string, parse.PrimaryType, *parse.Primary) {
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
		compound, current := primaryInSimpleCompound(primary)
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

func complIndexInner(m eval.IterateKeyer) []rawCandidate {
	var keys []rawCandidate
	m.IterateKey(func(v eval.Value) bool {
		if keyv, ok := v.(eval.String); ok {
			keys = append(keys, plainCandidate(keyv))
		}
		return true
	})
	sort.Sort(plainCandidates(keys))
	return keys
}

func complFormHead(n parse.Node, ev *eval.Evaler) (*compl, error) {
	begin, end, head, q := findFormHeadContext(n)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}
	cands, err := complFormHeadInner(head, ev)
	if err != nil {
		return nil, err
	}
	return &compl{begin, end, cookCandidates(cands, head, q)}, nil
}

func findFormHeadContext(n parse.Node) (int, int, string, parse.PrimaryType) {
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
		if compound, head := primaryInSimpleCompound(primary); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return compound.Begin(), compound.End(), head, primary.Type
				}
			}
		}
	}
	return -1, -1, "", 0
}

func complFormHeadInner(head string, ev *eval.Evaler) ([]rawCandidate, error) {
	if util.DontSearch(head) {
		return complFilenameInner(head, true)
	}

	var commands []rawCandidate
	got := func(s string) {
		commands = append(commands, plainCandidate(s))
	}
	for special := range isBuiltinSpecial {
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
	for ns := range ev.Modules {
		if head != ns+":" {
			got(ns + ":")
		}
	}
	sort.Sort(plainCandidates(commands))

	return commands, nil
}

type plainCandidates []rawCandidate

func (pc plainCandidates) Len() int { return len(pc) }
func (pc plainCandidates) Less(i, j int) bool {
	return pc[i].(plainCandidate) < pc[j].(plainCandidate)
}
func (pc plainCandidates) Swap(i, j int) { pc[i], pc[j] = pc[j], pc[i] }

// complRedir completes redirection RHS.
func complRedir(n parse.Node, ev *eval.Evaler) (*compl, error) {
	begin, end, current, q := findRedirContext(n)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}
	cands, err := complFilenameInner(current, false)
	if err != nil {
		return nil, err
	}
	return &compl{begin, end, cookCandidates(cands, current, q)}, nil
}

func findRedirContext(n parse.Node) (int, int, string, parse.PrimaryType) {
	if parse.IsSep(n) {
		if parse.IsRedir(n.Parent()) {
			return n.End(), n.End(), "", parse.Bareword
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary); compound != nil {
			if parse.IsRedir(compound.Parent()) {
				return compound.Begin(), compound.End(), head, primary.Type
			}
		}
	}
	return -1, -1, "", 0
}

// complArg completes arguments. It identifies the context and then delegates
// the actual completion work to a suitable completer.
func complArg(n parse.Node, ev *eval.Evaler) (*compl, error) {
	begin, end, current, q, form := findArgContext(n)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}

	// Find out head of the form and preceding arguments.
	// If Form.Head is not a simple compound, head will be "", just what we want.
	_, head, _ := simpleCompound(form.Head, nil)
	var args []string
	for _, compound := range form.Args {
		if compound.Begin() >= begin {
			break
		}
		ok, arg, _ := simpleCompound(compound, nil)
		if ok {
			// XXX Arguments that are not simple compounds are simply ignored.
			args = append(args, arg)
		}
	}

	words := make([]string, len(args)+2)
	words[0] = head
	words[len(words)-1] = current
	copy(words[1:len(words)-1], args[:])

	cands, err := completeArg(words, ev)
	if err != nil {
		return nil, err
	}
	return &compl{begin, end, cookCandidates(cands, current, q)}, nil
}

func findArgContext(n parse.Node) (int, int, string, parse.PrimaryType, *parse.Form) {
	if sep, ok := n.(*parse.Sep); ok {
		if form, ok := sep.Parent().(*parse.Form); ok {
			return n.End(), n.End(), "", parse.Bareword, form
		}
	}
	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head != compound {
					return compound.Begin(), compound.End(), head, primary.Type, form
				}
			}
		}
	}
	return -1, -1, "", 0, nil
}

// TODO: getStyle does redundant stats.
func complFilenameInner(head string, executableOnly bool) (
	[]rawCandidate, error) {

	dir, fileprefix := path.Split(head)
	dirToRead := dir
	if dirToRead == "" {
		dirToRead = "."
	}

	infos, err := ioutil.ReadDir(dirToRead)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory %s: %v", dirToRead, err)
	}

	cands := []rawCandidate{}
	lsColor := getLsColor()
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
			suffix = "/"
		} else if info.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(full)
			if err == nil && stat.IsDir() {
				// Symlink to directory.
				suffix = "/"
			}
		}

		cands = append(cands, &complexCandidate{
			stem: full, codeSuffix: suffix,
			style: stylesFromString(lsColor.getStyle(full)),
		})
	}

	return cands, nil
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
