// Package parseutil contains utilities built on top of the parse package.
package parseutil

import (
	"strings"

	"src.elv.sh/pkg/parse"
)

// FindLeafNode finds the leaf node at a specific position. It returns nil if
// position is out of bound.
func FindLeafNode(n parse.Node, p int) parse.Node {
descend:
	for len(parse.Children(n)) > 0 {
		for _, ch := range parse.Children(n) {
			if ch.Range().From <= p && p <= ch.Range().To {
				n = ch
				continue descend
			}
		}
		return nil
	}
	return n
}

// Wordify turns a piece of source code into words.
func Wordify(src string) []string {
	tree, _ := parse.Parse(parse.Source{Code: src}, parse.Config{})
	return wordifyInner(tree.Root, nil)
}

func wordifyInner(n parse.Node, words []string) []string {
	if len(parse.Children(n)) == 0 || isCompound(n) {
		text := parse.SourceText(n)
		if strings.TrimFunc(text, parse.IsWhitespace) != "" {
			return append(words, text)
		}
		return words
	}
	for _, ch := range parse.Children(n) {
		words = wordifyInner(ch, words)
	}
	return words
}

func isCompound(n parse.Node) bool {
	_, ok := n.(*parse.Compound)
	return ok
}
