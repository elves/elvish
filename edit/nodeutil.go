package edit

import (
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Utilities for insepcting the AST. Used for completers and stylists.

func primaryInSimpleCompound(pn *parse.Primary) (*parse.Compound, string) {
	thisIndexing, ok := pn.Parent().(*parse.Indexing)
	if !ok {
		return nil, ""
	}
	thisCompound, ok := thisIndexing.Parent().(*parse.Compound)
	if !ok {
		return nil, ""
	}
	ok, head, _ := simpleCompound(thisCompound, thisIndexing)
	if !ok {
		return nil, ""
	}
	return thisCompound, head
}

func simpleCompound(cn *parse.Compound, upto *parse.Indexing) (bool, string, error) {
	tilde := false
	head := ""
	for _, in := range cn.Indexings {
		if len(in.Indicies) > 0 {
			return false, "", nil
		}
		switch in.Head.Type {
		case parse.Tilde:
			tilde = true
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			head += in.Head.Value
		default:
			return false, "", nil
		}

		if in == upto {
			break
		}
	}
	if tilde {
		i := strings.Index(head, "/")
		if i == -1 {
			i = len(head)
		}
		uname := head[:i]
		home, err := util.GetHome(uname)
		if err != nil {
			return false, "", err
		}
		head = home + head[i:]
	}
	return true, head, nil
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
	if len(n.Children()) == 0 {
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
