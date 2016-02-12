package edit

import (
	"strings"

	"github.com/elves/elvish/osutil"
	"github.com/elves/elvish/parse"
)

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

	tilde := false
	head := ""
	for _, in := range thisCompound.Indexings {
		if len(in.Indicies) > 0 {
			return nil, ""
		}
		switch in.Head.Type {
		case parse.Tilde:
			tilde = true
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			head += in.Head.Value
		}

		if in == thisIndexing {
			break
		}
	}
	if tilde {
		i := strings.Index(head, "/")
		var home string
		var err error
		if i == -1 {
			home, err = osutil.GetHome(head)
			head = home
		} else {
			uname := head[:i]
			home, err = osutil.GetHome(uname)
			head = home + head[i:]
		}
		if err != nil {
			// TODO report error
			return nil, ""
		}
	}
	return thisCompound, head
}
