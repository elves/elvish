// Package parseutil contains utilities built on top of the parse package.
package parseutil

import (
	"strings"

	"src.elv.sh/pkg/parse"
)

// Wordify turns a piece of source code into words.
func Wordify(src string) []string {
	tree, _ := parse.Parse(parse.Source{Name: "[unknown]", Code: src}, parse.Config{})
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
