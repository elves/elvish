package edit

//go:generate stringer -type=TokenKind

import (
	"strings"

	"github.com/elves/elvish/parse"
)

var tokensBufferSize = 16

// Token is a leaf of the parse tree.
type Token struct {
	Type      TokenKind
	Text      string
	Node      parse.Node
	MoreStyle styles
}

// TokenKind classifies Token's.
type TokenKind int

// Values for TokenKind.
const (
	ParserError TokenKind = iota
	Bareword
	SingleQuoted
	DoubleQuoted
	Variable
	Wildcard
	Tilde
	Sep
)

func parserError(src string, begin, end int) Token {
	return Token{ParserError, src[begin:end], parse.NewSep(src, begin, end), styles{}}
}

// tokenize tokenizes elvishscript source.
func tokenize(src string) []Token {
	node, _ := parse.Parse("[tokenize]", src)
	return tokenizeNode(src, node)
}

// wordify breaks elvishscript source into words by breaking it into tokens, and
// pick out the texts of non-space non-newline tokens.
func wordify(src string) []string {
	tokens := tokenize(src)
	var words []string
	for _, token := range tokens {
		text := token.Text
		if strings.TrimFunc(text, parse.IsSpaceOrNewline) != "" {
			words = append(words, text)
		}
	}
	return words
}

// tokenizeNode returns all leaves in an AST.
func tokenizeNode(src string, n parse.Node) []Token {
	lastEnd := 0

	tokenCh := make(chan Token, tokensBufferSize)
	tokens := []Token{}
	tokensDone := make(chan bool)

	go func() {
		for token := range tokenCh {
			begin := token.Node.Begin()
			if begin > lastEnd {
				tokens = append(tokens, parserError(src, lastEnd, begin))
			}
			tokens = append(tokens, token)
			lastEnd = token.Node.End()
		}
		tokensDone <- true
	}()
	produceTokens(n, tokenCh)
	close(tokenCh)

	<-tokensDone
	if lastEnd != len(src) {
		tokens = append(tokens, parserError(src, lastEnd, len(src)))
	}
	return tokens
}

func produceTokens(n parse.Node, tokenCh chan<- Token) {
	if n.Begin() == n.End() {
		// Ignore empty node. This happens e.g. with an empty source code, where
		// the parsed node is an empty Chunk.
		return
	}
	if len(n.Children()) == 0 {
		tokenType := ParserError
		moreStyle := styles{}
		switch n := n.(type) {
		case *parse.Primary:
			switch n.Type {
			case parse.Bareword:
				tokenType = Bareword
			case parse.SingleQuoted:
				tokenType = SingleQuoted
			case parse.DoubleQuoted:
				tokenType = DoubleQuoted
			case parse.Variable:
				tokenType = Variable
			case parse.Wildcard:
				tokenType = Wildcard
			case parse.Tilde:
				tokenType = Tilde
			}
		case *parse.Sep:
			tokenType = Sep
			septext := n.SourceText()
			if strings.HasPrefix(septext, "#") {
				moreStyle = append(moreStyle, styleForSep["#"])
			} else {
				moreStyle = append(moreStyle, styleForSep[septext])
			}
		default:
			Logger.Printf("bad leaf type %T", n)
		}
		tokenCh <- Token{tokenType, n.SourceText(), n, moreStyle}
	}
	for _, child := range n.Children() {
		produceTokens(child, tokenCh)
	}
}
