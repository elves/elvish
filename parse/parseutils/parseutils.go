// Package parseutils contains utilities built on top of the parse package.
package parseutils

import (
	"strings"

	"github.com/elves/elvish/parse"
)

// Wordify turns a piece of source code into words.
func Wordify(src string) []string {
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
