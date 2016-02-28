package edit

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// A completer takes the current Node and an Editor and returns an interval and
// a list of candidates, meaning that the text within the interval may be
// replaced by any of the candidates. the Node is always a leaf in the parsed
// AST.
type completer func(parse.Node, *Editor) (int, int, []*candidate)

var completers = []struct {
	name string
	completer
}{
	{"variable", complVariable},
	{"command name", complFormHead},
	{"argument", complArg},
}

func complVariable(n parse.Node, ed *Editor) (begin, end int, cands []*candidate) {
	begin, end = n.Begin(), n.End()

	primary, ok := n.(*parse.Primary)
	if !ok || primary.Type != parse.Variable {
		return
	}

	head := primary.Value

	// Collect matching variables.
	var varnames []string
	for varname := range eval.Builtin() {
		if strings.HasPrefix(varname, head) {
			varnames = append(varnames, varname)
		}
	}
	for varname := range ed.evaler.Global() {
		if strings.HasPrefix(varname, head) {
			varnames = append(varnames, varname)
		}
	}
	sort.Strings(varnames)

	// Build candidates.
	for _, varname := range varnames {
		cands = append(cands, &candidate{
			source: styled{"$" + varname, styleForType[Variable]},
			menu:   styled{"$" + varname, ""}})
	}
	return
}

func complFormHead(n parse.Node, ed *Editor) (int, int, []*candidate) {
	begin, end, head, q := findFormHeadContext(n)
	if begin == -1 {
		return -1, -1, nil
	}
	cands, err := complFormHeadInner(head, ed)
	if err != nil {
		ed.notify("%v", err)
	}
	fixCandidates(cands, q)
	return begin, end, cands
}

func findFormHeadContext(n parse.Node) (int, int, string, parse.PrimaryType) {
	_, isChunk := n.(*parse.Chunk)
	_, isPipeline := n.(*parse.Pipeline)
	if isChunk || isPipeline {
		return n.Begin(), n.End(), "", parse.Bareword
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

func complFormHeadInner(head string, ed *Editor) ([]*candidate, error) {
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
	for variable := range eval.Builtin() {
		if strings.HasPrefix(variable, eval.FnPrefix) {
			got(variable[len(eval.FnPrefix):])
		}
	}
	for variable := range ed.evaler.Global() {
		if strings.HasPrefix(variable, eval.FnPrefix) {
			got(variable[len(eval.FnPrefix):])
		}
	}
	for command := range ed.isExternal {
		got(command)
	}
	sort.Strings(commands)

	cands := []*candidate{}
	for _, cmd := range commands {
		cands = append(cands, &candidate{
			source: styled{cmd, styleForGoodCommand},
			menu:   styled{cmd, ""}})
	}
	return cands, nil
}

func complArg(n parse.Node, ed *Editor) (int, int, []*candidate) {
	begin, end, current, q, form := findArgContext(n)
	if begin == -1 {
		return 0, 0, nil
	}

	// Find out head of the form and preceding arguments.
	// If Form.Head is not a simple compound, head will be "", just what we want.
	_, head := simpleCompound(form.Head, nil)
	var args []string
	for _, compound := range form.Args {
		if compound.Begin() >= begin {
			break
		}
		ok, arg := simpleCompound(compound, nil)
		if ok {
			// XXX Arguments that are not simple compounds are simply ignored.
			args = append(args, arg)
		}
	}

	cands, err := completeArg(&ArgContext{head, args, current, ed})
	if err != nil {
		ed.notify("%v", err)
	}
	fixCandidates(cands, q)
	return begin, end, cands
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
		}

		cands = append(cands, &candidate{
			source: styled{full, ""}, sourceSuffix: suffix,
			menu: styled{name, defaultLsColor.getStyle(full)},
		})
	}

	return cands, nil
}

func fixCandidates(cands []*candidate, q parse.PrimaryType) []*candidate {
	for _, cand := range cands {
		quoted, _ := parse.QuoteAs(cand.source.text, q)
		cand.source.text = quoted + cand.sourceSuffix
	}
	return cands
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
