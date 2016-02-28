package edit

import (
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
	{"command name", complNewForm},
	{"command name", makeCompoundCompleter(complFormHead)},
	{"argument", complNewArg},
	{"argument", makeCompoundCompleter(complArg)},
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

func complNewForm(n parse.Node, ed *Editor) (begin, end int, cands []*candidate) {
	begin, end = n.Begin(), n.End()
	if _, ok := n.(*parse.Chunk); ok {
		// Leaf is a chunk is a leaf. Must be empty chunk.
		cands = complFormHeadInner("", ed)
		return
	}
	if _, ok := n.Parent().(*parse.Chunk); ok {
		// Parent is a chunk. Must be an empty pipeline.
		cands = complFormHeadInner("", ed)
		return
	}
	return
}

// XXX Semantics of candidate is pretty broken here.
func makeCompoundCompleter(
	f func(*parse.Compound, string, *Editor) []*candidate) completer {
	return func(n parse.Node, ed *Editor) (begin, end int, cands []*candidate) {
		begin, end = n.Begin(), n.End()

		pn, ok := n.(*parse.Primary)
		if !ok {
			return
		}
		cn, head := simpleCompound(pn)
		if cn == nil {
			return
		}
		cands = f(cn, head, ed)
		for _, cand := range cands {
			quoted, _ := parse.QuoteAs(cand.source.text, pn.Type)
			cand.source.text = quoted + cand.sourceSuffix
		}
		return
	}
}

func complFormHead(cn *parse.Compound, head string, ed *Editor) []*candidate {
	if isFormHead(cn) {
		return complFormHeadInner(head, ed)
	}
	return nil
}

func complFormHeadInner(head string, ed *Editor) []*candidate {
	if util.DontSearch(head) {
		return complArgInner(head, ed, true)
	}

	cands := []*candidate{}

	foundCommand := func(s string) {
		if strings.HasPrefix(s, head) {
			cands = append(cands, &candidate{
				source: styled{s, styleForGoodCommand},
				menu:   styled{s, ""},
			})
		}
	}
	for special := range isBuiltinSpecial {
		foundCommand(special)
	}
	for variable := range eval.Builtin() {
		if strings.HasPrefix(variable, eval.FnPrefix) {
			foundCommand(variable[len(eval.FnPrefix):])
		}
	}
	for variable := range ed.evaler.Global() {
		if strings.HasPrefix(variable, eval.FnPrefix) {
			foundCommand(variable[len(eval.FnPrefix):])
		}
	}
	for command := range ed.isExternal {
		foundCommand(command)
	}
	return cands
}

func complNewArg(n parse.Node, ed *Editor) (begin int, end int, cands []*candidate) {
	begin, end = n.End(), n.End()
	sn, ok := n.(*parse.Sep)
	if !ok {
		return
	}
	if _, ok := sn.Parent().(*parse.Form); !ok {
		return
	}
	cands = complArgInner("", ed, false)
	return
}

func complArg(cn *parse.Compound, head string, ed *Editor) []*candidate {
	return complArgInner(head, ed, false)
}

// TODO: getStyle does redundant stats.
func complArgInner(head string, ed *Editor, formHead bool) []*candidate {
	dir, fileprefix := path.Split(head)
	if dir == "" {
		dir = "."
	}

	infos, err := ioutil.ReadDir(dir)
	cands := []*candidate{}

	if err != nil {
		ed.addTip("cannot list directory %s: %v", dir, err)
		return cands
	}

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
		// completing head.
		if formHead && !(info.IsDir() || (info.Mode()&0111) != 0) {
			continue
		}

		// Full filename for .getStyle.
		full := head + name[len(fileprefix):]

		var suffix string
		if info.IsDir() {
			suffix = "/"
		} else {
			suffix = " "
		}

		cands = append(cands, &candidate{
			source:       styled{name, ""},
			menu:         styled{name, defaultLsColor.getStyle(full)},
			sourceSuffix: suffix,
		})
	}

	return cands
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}
