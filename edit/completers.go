package edit

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// A completer takes the current node
type completer func(parse.Node) []*candidate

var completers = []struct {
	name string
	completer
}{
	{"command name", complFormHead},
	{"argument", complArg},
}

func complFormHead(n parse.Node) []*candidate {
	if !isFormHead(n) {
		return nil
	}
	head := n.SourceText()

	cands := []*candidate{}
	for _, s := range builtins {
		if strings.HasPrefix(s, head) {
			cands = append(cands, newCandidate(
				tokenPart{head, false}, tokenPart{s[len(head):], true}))
		}
	}
	return cands
}

var builtins []string

func init() {
	builtins = append(builtins, eval.BuiltinFnNames...)
	builtins = append(builtins, eval.BuiltinSpecialNames...)
}

func isFormHead(n parse.Node) bool {
	if _, ok := n.(*parse.Chunk); ok {
		return true
	}
	if n, ok := n.(*parse.Primary); ok {
		if n, ok := n.Parent().(*parse.Indexed); ok {
			if compound, ok := n.Parent().(*parse.Compound); ok {
				if form, ok := compound.Parent().(*parse.Form); ok {
					return compound == form.Head
				}
			}
		}
	}
	return false
}

func complArg(n parse.Node) []*candidate {
	head := n.SourceText()

	// Assume that the token is an incomplete filename
	dir, file := path.Split(head)
	var all []string
	if dir == "" {
		// XXX ignore error
		all, _ = fileNames(".")
	}
	all, _ = fileNames(dir)

	cands := []*candidate{}
	// Make candidates out of elements that match the file component.
	for _, s := range all {
		if strings.HasPrefix(s, file) {
			cand := newCandidate(
				tokenPart{head, false}, tokenPart{s[len(file):], true})
			cand.attr = defaultLsColor.determineAttr(cand.text)
			cands = append(cands, cand)
		}
	}
	return cands
}

func fileNames(dir string) (names []string, err error) {
	infos, e := ioutil.ReadDir(dir)
	if e != nil {
		err = e
		return
	}
	for _, info := range infos {
		names = append(names, info.Name())
	}
	return
}
