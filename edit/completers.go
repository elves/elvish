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

// completer takes the current Node (always a leaf in the AST) and an Editor and
// returns a compl. If the completer does not apply to the type of the current
// Node, it should return an error of ErrCompletionUnapplicable.
type completer func(parse.Node, *eval.Evaler) (*compl, error)

// compl is the result of a completer, meaning that any of the candidates can
// replace the text in the interval [begin, end).
type compl struct {
	begin int
	end   int
	cands []*candidate
}

var errCompletionUnapplicable = errors.New("completion unapplicable")

// completers is the list of all completers.
// TODO(xiaq): Make this list programmable.
var completers = []struct {
	name string
	completer
}{
	{"variable", complVariable},
	{"command name", complFormHead},
	{"redir", complRedir},
	{"argument", complArg},
}

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
		cands[i] = newPlainCandidate("$" + explode + varname)
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

func complFormHead(n parse.Node, ev *eval.Evaler) (*compl, error) {
	begin, end, head, q := findFormHeadContext(n)
	if begin == -1 {
		return nil, errCompletionUnapplicable
	}
	cands, err := complFormHeadInner(head, ev)
	if err != nil {
		return nil, err
	}
	quoteCandidates(cands, q)
	return &compl{begin, end, cands}, nil
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

func complFormHeadInner(head string, ev *eval.Evaler) ([]*candidate, error) {
	if util.DontSearch(head) {
		return complFilenameInner(head, true)
	}

	var commands []string
	got := func(s string) {
		if strings.HasPrefix(s, head) {
			commands = append(commands, s)
		}
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
	sort.Strings(commands)

	cands := []*candidate{}
	for _, cmd := range commands {
		cands = append(cands, newPlainCandidate(cmd))
	}
	return cands, nil
}

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
	quoteCandidates(cands, q)
	return &compl{begin, end, cands}, nil
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
	quoteCandidates(cands, q)
	return &compl{begin, end, cands}, nil
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
func complFilenameInner(head string, executableOnly bool) ([]*candidate, error) {
	dir, fileprefix := path.Split(head)
	if dir == "" {
		dir = "."
	}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot list directory %s: %v", dir, err)
	}

	cands := []*candidate{}
	lsColor := getLsColor()
	// Make candidates out of elements that match the file component.
	for _, info := range infos {
		name := info.Name()
		// Irrevelant file.
		if !strings.HasPrefix(name, fileprefix) {
			continue
		}
		// Hide dot files unless file starts with a dot.
		if !dotfile(fileprefix) && dotfile(name) {
			continue
		}
		// Only accept searchable directories and executable files if
		// executableOnly is true.
		if executableOnly && !(info.IsDir() || (info.Mode()&0111) != 0) {
			continue
		}

		// Full filename for source and getStyle.
		full := head + name[len(fileprefix):]

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

		cands = append(cands, &candidate{
			text: full, suffix: suffix,
			display: styled{name, stylesFromString(lsColor.getStyle(full))},
		})
	}

	return cands, nil
}

func quoteCandidates(cands []*candidate, q parse.PrimaryType) []*candidate {
	for _, cand := range cands {
		quoted, _ := parse.QuoteAs(cand.text, q)
		cand.text = quoted + cand.suffix
	}
	return cands
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
