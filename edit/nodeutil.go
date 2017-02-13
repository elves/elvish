package edit

import (
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Utilities for insepcting the AST. Used for completers and stylists.

func primaryInSimpleCompound(pn *parse.Primary) (*parse.Compound, string) {
	indexing := parse.GetIndexing(pn.Parent())
	if indexing == nil {
		return nil, ""
	}
	compound := parse.GetCompound(indexing.Parent())
	if compound == nil {
		return nil, ""
	}
	ok, head, _ := simpleCompound(compound, indexing)
	if !ok {
		return nil, ""
	}
	return compound, head
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

// purelyEvalPrimary evaluates a primary node without causing any side effects.
// If this cannot be done, it returns nil.
//
// Currently, only string literals and variables with no @ can be evaluated.
func purelyEvalPrimary(pn *parse.Primary, ev *eval.Evaler) eval.Value {
	switch pn.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return eval.String(pn.Value)
	case parse.Variable:
		explode, ns, name := eval.ParseVariable(pn.Value)
		if explode {
			return nil
		}
		ec := eval.NewTopEvalCtx(ev, "[pure eval]", "", nil)
		variable := ec.ResolveVar(ns, name)
		return variable.Get()
	}
	return nil
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
