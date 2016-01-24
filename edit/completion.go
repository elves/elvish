package edit

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse-ng"
)

type tokenPart struct {
	text      string
	completed bool
}

type candidate struct {
	text  string
	parts []tokenPart
	attr  string // Attribute used for preview
}

func newCandidate() *candidate {
	return &candidate{}
}

func (c *candidate) push(tp tokenPart) {
	c.text += tp.text
	c.parts = append(c.parts, tp)
}

type completion struct {
	start, end int // The text to complete is Editor.line[start:end]
	typ        TokenType
	candidates []*candidate
	current    int
}

func (c *completion) prev(cycle bool) {
	c.current--
	if c.current == -1 {
		if cycle {
			c.current = len(c.candidates) - 1
		} else {
			c.current++
		}
	}
}

func (c *completion) next(cycle bool) {
	c.current++
	if c.current == len(c.candidates) {
		if cycle {
			c.current = 0
		} else {
			c.current--
		}
	}
}

// Find candidates by matching against a prefix.
func prefixMatchCandidates(p string, all []string) (cands []*candidate) {
	for _, s := range all {
		if len(s) >= len(p) && s[:len(p)] == p {
			cand := newCandidate()
			cand.push(tokenPart{p, false})
			cand.push(tokenPart{s[len(p):], true})
			cands = append(cands, cand)
		}
	}
	return
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

var builtins []string

func init() {
	builtins = append(builtins, eval.BuiltinFnNames...)
	builtins = append(builtins, eval.BuiltinSpecialNames...)
}

func startCompletion(ed *Editor, k Key) *leReturn {
	c := &completion{}
	token := tokenAtDot(ed)
	node := token.Node
	var findAll func() ([]string, error)
	var makeCandidates func([]string) []*candidate

	head := token.Text

	ed.pushTip(fmt.Sprintf("current node is %T", node))
	if isFormHead(node) {
		// BUG(xiaq): When completing commands, only builtins are searched
		findAll = func() ([]string, error) {
			return builtins, nil
		}
		makeCandidates = func(all []string) (cands []*candidate) {
			return prefixMatchCandidates(head, all)
		}
	} else {
		// Assume that the token is an incomplete filename
		dir, file := path.Split(head)
		findAll = func() ([]string, error) {
			if dir == "" {
				return fileNames(".")
			}
			return fileNames(dir)
		}
		makeCandidates = func(all []string) (cands []*candidate) {
			// Make candidates out of elements that match the file component.
			for _, s := range all {
				if strings.HasPrefix(s, file) {
					cand := newCandidate()
					cand.push(tokenPart{head, false})
					cand.push(tokenPart{s[len(file):], true})
					cand.attr = defaultLsColor.determineAttr(cand.text)
					cands = append(cands, cand)
				}
			}
			return
		}
	}

	// BUG(xiaq): When completing, only plain expressions are supported
	all, err := findAll()
	if err != nil {
		ed.pushTip(err.Error())
		return nil
	}
	c.start = node.N().Begin
	c.end = ed.dot
	// BUG(xiaq) When completing, completion.typ is always ItemBare
	c.typ = Bareword
	c.candidates = makeCandidates(all)
	if len(c.candidates) > 0 {
		ed.completion = c
		ed.mode = modeCompletion
	} else {
		ed.pushTip(fmt.Sprintf("No completion for %s", head))
	}
	return nil
}

var BadToken = Token{}

func tokenAtDot(ed *Editor) Token {
	if len(ed.tokens) == 0 || ed.dot > len(ed.line) {
		return BadToken
	}
	if ed.dot == len(ed.line) {
		return ed.tokens[len(ed.tokens)-1]
	}
	for _, token := range ed.tokens {
		if ed.dot < token.Node.N().End {
			return token
		}
	}
	return BadToken
}

func isFormHead(n parse.Node) bool {
	if _, ok := n.(*parse.Chunk); ok {
		return true
	}
	if n, ok := n.(*parse.Primary); ok {
		if n, ok := n.N().Parent.(*parse.Indexed); ok {
			if n, ok := n.N().Parent.(*parse.Compound); ok {
				if _, ok := n.N().Parent.(*parse.Form); ok {
					return true
				}
			}
		}
	}
	return false
}
