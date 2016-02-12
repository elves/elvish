package edit

import "github.com/elves/elvish/parse"

var tokensBufferSize = 16

// Token is a leaf of the parse tree.
type Token struct {
	Type      TokenKind
	Text      string
	Node      parse.Node
	MoreStyle string
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

func parserError(text string) Token {
	return Token{ParserError, text, nil, ""}
}

// tokenize returns all leaves in an AST.
func tokenize(src string, n parse.Node) []Token {
	lastEnd := 0

	tokenCh := make(chan Token, tokensBufferSize)
	tokens := []Token{}
	tokensDone := make(chan bool)

	go func() {
		for token := range tokenCh {
			begin := token.Node.Begin()
			if begin > lastEnd {
				tokens = append(tokens, parserError(src[lastEnd:begin]))
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
		tokens = append(tokens, parserError(src[lastEnd:]))
	}
	return tokens
}

func produceTokens(n parse.Node, tokenCh chan<- Token) {
	if len(n.Children()) == 0 {
		tokenType := ParserError
		moreStyle := ""
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
			moreStyle = styleForSep[n.SourceText()]
		}
		tokenCh <- Token{tokenType, n.SourceText(), n, moreStyle}
	}
	for _, child := range n.Children() {
		produceTokens(child, tokenCh)
	}
}
