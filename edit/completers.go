package edit

import (
	"fmt"
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

	head := primary.Value[1:]
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
		return complArgInner(head, false, ed, true)
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
	return complArgInner("", false, ed, false)
}

func complArg(cn *parse.Compound, head string, ed *Editor) []*candidate {
	return complArgInner(head, false, ed, false)
}

// TODO: all of fileNames, getStyle and the final directory check do stat on
// files.
func complArgInner(head string, indir bool, ed *Editor, formHead bool) []*candidate {
	var dir, file, indirSlash string
	if indir {
		dir = head
		indirSlash = "/"
	} else {
		dir, file = path.Split(head)
		if dir == "" {
			dir = "."
		}
	}
	names, err := fileNames(dir)
	cands := []*candidate{}

	if err != nil {
		ed.pushTip(fmt.Sprintf("cannot list directory %s: %v", dir, err))
		return cands
	}

	hasHead := false
	// Make candidates out of elements that match the file component.
	for s := range names {
		if strings.HasPrefix(s, file) {
			if s == file {
				hasHead = true
			}
			full := head + indirSlash + s[len(file):]
			if formHead && !isExecutableOrDir(full) {
				continue
			}
			cand := &candidate{
				source: styled{indirSlash + s[len(file):], ""},
				menu:   styled{s, defaultLsColor.getStyle(full)},
			}
			cands = append(cands, cand)
		}
	}

	if !indir && hasHead && len(cands) == 1 && isDir(head) {
		// Completing an unambiguous directory name.
		return complArgInner(head, true, ed, formHead)
	}

	return cands
}

func isDir(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && stat.IsDir()
}

func isExecutableOrDir(fname string) bool {
	stat, err := os.Stat(fname)
	return err == nil && (stat.Mode()&0111 != 0)
}

func fileNames(dir string) (<-chan string, error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make(chan string, 32)
	go func() {
		for _, info := range infos {
			names <- info.Name()
		}
		close(names)
	}()
	return names, nil
}
