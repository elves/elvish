package edit

import (
	"strings"

	"github.com/elves/elvish/edit/nodeutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// Utilities for insepcting the AST. Used for completers and stylists.

func primaryInSimpleCompound(pn *parse.Primary, ev *eval.Evaler) (*parse.Compound, string) {
	indexing := parse.GetIndexing(pn.Parent())
	if indexing == nil {
		return nil, ""
	}
	compound := parse.GetCompound(indexing.Parent())
	if compound == nil {
		return nil, ""
	}
	ok, head, _ := simpleCompound(compound, indexing, ev)
	if !ok {
		return nil, ""
	}
	return compound, head
}

func simpleCompound(cn *parse.Compound, upto *parse.Indexing, ev *eval.Evaler) (bool, string, error) {
	return nodeutil.SimpleCompound(cn, upto, ev)
}

func purelyEvalPrimary(cn *parse.Primary, ev *eval.Evaler) eval.Value {
	return nodeutil.PurelyEvalPrimary(cn, ev)
}

// leafNodeAtDot finds the leaf node at a specific position. It returns nil if
// position is out of bound.
func findLeafNode(n parse.Node, p int) parse.Node {
descend:
	for len(n.Children()) > 0 {
		for _, ch := range n.Children() {
			if ch.Begin() <= p && p <= ch.End() {
				n = ch
				continue descend
			}
		}
		return nil
	}
	return n
}

func wordify(src string) []string {
	n, _ := parse.Parse("[wordify]", src)
	return wordifyInner(n, nil)
}

func wordifyInner(n parse.Node, words []string) []string {
	if len(n.Children()) == 0 || parse.IsCompound(n) {
		text := n.SourceText()
		if strings.TrimFunc(text, parse.IsSpaceOrNewline) != "" {
			return append(words, text)
		}
		return words
	}
	for _, ch := range n.Children() {
		words = wordifyInner(ch, words)
	}
	return words
}
