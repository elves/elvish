package edit

import "github.com/elves/elvish/parse"

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
		if compound, head := simpleCompound(primary); compound != nil {
			if form, ok := compound.Parent().(*parse.Form); ok {
				if form.Head == compound {
					return compound, head
				}
			}
		}
	}

	return nil, ""
}

func simpleCompound(pn *parse.Primary) (*parse.Compound, string) {
	thisIndexing, ok := pn.Parent().(*parse.Indexing)
	if !ok {
		return nil, ""
	}

	thisCompound, ok := thisIndexing.Parent().(*parse.Compound)
	if !ok {
		return nil, ""
	}

	head := ""
	for _, in := range thisCompound.Indexings {
		if len(in.Indicies) > 0 {
			return nil, ""
		}
		typ := in.Head.Type
		if typ != parse.Bareword &&
			typ != parse.SingleQuoted &&
			typ != parse.DoubleQuoted {
			return nil, ""
		}
		head += in.Head.Value
		if in == thisIndexing {
			break
		}
	}
	return thisCompound, head
}
