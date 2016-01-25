package edit

import "github.com/elves/elvish/parse"

func formHead(n parse.Node) (parse.Node, string) {
	if _, ok := n.(*parse.Chunk); ok {
		return n, ""
	}

	if primary, ok := n.(*parse.Primary); ok {
		compound, head := simpleCompound(primary)
		if form, ok := compound.Parent().(*parse.Form); ok {
			if form.Head == compound {
				return compound, head
			}
		}
	}

	return nil, ""
}

func simpleCompound(pn *parse.Primary) (*parse.Compound, string) {
	thisIndexed, ok := pn.Parent().(*parse.Indexed)
	if !ok {
		return nil, ""
	}

	thisCompound, ok := thisIndexed.Parent().(*parse.Compound)
	if !ok {
		return nil, ""
	}

	head := ""
	for _, in := range thisCompound.Indexeds {
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
		if in == thisIndexed {
			break
		}
	}
	return thisCompound, head
}
