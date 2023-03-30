// Package parseutil contains utilities built on top of the parse package.
package parseutil

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/parse"
)

// Wordify turns a piece of source code into words.
func Wordify(src string) []string {
	tree, _ := parse.Parse(parse.Source{Name: "[unknown]", Code: src}, parse.Config{})
	return wordifyInner(tree.Root, nil)
}

func LeafTextAtPos(n parse.Node, pos int) (string, error) {
	if len(parse.Children(n)) == 0 {
		text := parse.SourceText(n)
		rnge := n.Range()
		if pos >= rnge.From && pos < rnge.To {
			return text, nil
		}
		return "", fmt.Errorf("no leaf parse.Node at document pos %d found", pos)
	}
	for _, ch := range parse.Children(n) {
		text, err := LeafTextAtPos(ch, pos)
		if err == nil {
			return text, err
		}
	}
	return "", fmt.Errorf("no leaf parse.Node at document pos %d found", pos)
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
