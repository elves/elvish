package edit

import (
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Utilities for insepcting the AST. Used for completers and stylists.

func isFormHead(compound *parse.Compound) bool {
	if form, ok := compound.Parent().(*parse.Form); ok {
		return form.Head == compound
	}
	return false
}

func formHead(n parse.Node) (parse.Node, string) {
	if _, ok := n.(*parse.Chunk); ok {
		return n, ""
	}

	if primary, ok := n.(*parse.Primary); ok {
		if compound, head := primaryInSimpleCompound(primary); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return compound, head
				}
			}
		}
	}

	return nil, ""
}

func primaryInSimpleCompound(pn *parse.Primary) (*parse.Compound, string) {
	thisIndexing, ok := pn.Parent().(*parse.Indexing)
	if !ok {
		return nil, ""
	}
	thisCompound, ok := thisIndexing.Parent().(*parse.Compound)
	if !ok {
		return nil, ""
	}
	ok, head := simpleCompound(thisCompound, thisIndexing)
	if !ok {
		return nil, ""
	}
	return thisCompound, head
}

func simpleCompound(cn *parse.Compound, upto *parse.Indexing) (bool, string) {
	tilde := false
	head := ""
	for _, in := range cn.Indexings {
		if len(in.Indicies) > 0 {
			return false, ""
		}
		switch in.Head.Type {
		case parse.Tilde:
			tilde = true
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			head += in.Head.Value
		}

		if in == upto {
			break
		}
	}
	if tilde {
		i := strings.Index(head, "/")
		if i == -1 {
			return false, ""
		}
		uname := head[:i]
		home, err := util.GetHome(uname)
		if err != nil {
			// TODO report error
			return false, ""
		}
		head = home + head[i:]
	}
	return true, head
}
