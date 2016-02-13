package edit

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// A completer takes the current node
type completer func(parse.Node, *Editor) []*candidate

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

func complVariable(n parse.Node, ed *Editor) []*candidate {
	primary, ok := n.(*parse.Primary)
	if !ok || primary.Type != parse.Variable {
		return nil
	}

	head := primary.Value
	cands := []*candidate{}
	for variable := range ed.evaler.Global() {
		if strings.HasPrefix(variable, head) {
			cands = append(cands, &candidate{
				source: styled{variable[len(head):], styleForType[Variable]},
				menu:   styled{"$" + variable, styleForType[Variable]}})
		}
	}
	return cands
}

func complNewForm(n parse.Node, ed *Editor) []*candidate {
	if _, ok := n.(*parse.Chunk); ok {
		return complFormHeadInner("", ed)
	}
	if _, ok := n.Parent().(*parse.Chunk); ok {
		return complFormHeadInner("", ed)
	}
	return nil
}

func makeCompoundCompleter(
	f func(*parse.Compound, string, *Editor) []*candidate) completer {
	return func(n parse.Node, ed *Editor) []*candidate {
		pn, ok := n.(*parse.Primary)
		if !ok {
			return nil
		}
		cn, head := simpleCompound(pn)
		if cn == nil {
			return nil
		}
		return f(cn, head, ed)
	}
}

func complFormHead(cn *parse.Compound, head string, ed *Editor) []*candidate {
	if isFormHead(cn) {
		return complFormHeadInner(head, ed)
	}
	return nil
}

func complFormHeadInner(head string, ed *Editor) []*candidate {
	if eval.DontSearch(head) {
		return complArgInner(head, ed, true)
	}

	cands := []*candidate{}

	foundCommand := func(s string) {
		if strings.HasPrefix(s, head) {
			cands = append(cands, &candidate{
				source: styled{s[len(head):], styleForGoodCommand},
				menu:   styled{s, ""},
			})
		}
	}
	for special := range isBuiltinSpecial {
		foundCommand(special)
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

func complNewArg(n parse.Node, ed *Editor) []*candidate {
	sn, ok := n.(*parse.Sep)
	if !ok {
		return nil
	}
	if _, ok := sn.Parent().(*parse.Form); !ok {
		return nil
	}
	return complArgInner("", ed, false)
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

		if info.IsDir() {
			name += "/"
		} else {
			name += " "
		}

		cands = append(cands, &candidate{
			source: styled{name[len(fileprefix):], ""},
			menu:   styled{name, defaultLsColor.getStyle(full)},
		})
	}

	return cands
}

func dotfile(fname string) bool {
	return strings.HasPrefix(fname, ".")
}

func isDir(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && stat.IsDir()
}
