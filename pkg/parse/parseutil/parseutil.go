// Package parseutil contains utilities built on top of the parse package.
package parseutil

import (
	"strings"

	"github.com/elves/elvish/pkg/parse"
)

// LeafNodeAtDot finds the leaf node at a specific position. It returns nil if
// position is out of bound.
func FindLeafNode(n parse.Node, p int) parse.Node {
descend:
	for len(n.Children()) > 0 {
		for _, ch := range n.Children() {
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
	n, _ := parse.AsChunk("[wordify]", src)
	return wordifyInner(n, nil)
}

func wordifyInner(n parse.Node, words []string) []string {
	if len(n.Children()) == 0 || isCompound(n) {
		text := n.SourceText()
		if strings.TrimFunc(text, parse.IsWhitespace) != "" {
			return append(words, text)
		}
		return words
	}
	for _, ch := range n.Children() {
		words = wordifyInner(ch, words)
	}
	return words
}

func isCompound(n parse.Node) bool {
	_, ok := n.(*parse.Compound)
	return ok
}
