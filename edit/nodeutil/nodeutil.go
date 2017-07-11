// Package nodeutil provides utilities for inspecting the AST.
package nodeutil

import (
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func SimpleCompound(cn *parse.Compound, upto *parse.Indexing) (bool, string, error) {
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
