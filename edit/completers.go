package edit

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// A completer takes the current node
type completer func(parse.Node, *Editor) ([]*candidate, int)

var completers = []struct {
	name string
	completer
}{
	{"command name", complFormHead},
	{"argument", complArg},
}

func complFormHead(n parse.Node, ed *Editor) ([]*candidate, int) {
	n, head := formHead(n)
	if n == nil {
		return nil, 0
	}

	cands := []*candidate{}
	for _, s := range builtins {
		if strings.HasPrefix(s, head) {
			cands = append(cands, newCandidate(
				tokenPart{head, false}, tokenPart{s[len(head):], true}))
		}
	}
	return cands, n.Begin()
}

var builtins []string

func init() {
	builtins = append(builtins, eval.BuiltinFnNames...)
	builtins = append(builtins, eval.BuiltinSpecialNames...)
}

func complArg(n parse.Node, ed *Editor) ([]*candidate, int) {
	pn, ok := n.(*parse.Primary)
	if !ok {
		return nil, 0
	}
	cn, head := simpleCompound(pn)
	if cn == nil {
		return nil, 0
	}

	// Assume that the argument is an incomplete filename
	dir, file := path.Split(head)
	var all []string
	if dir == "" {
		// XXX ignore error
		all, _ = fileNames(".")
	} else {
		all, _ = fileNames(dir)
	}

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

	return cands, cn.Begin()
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
