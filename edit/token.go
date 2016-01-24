package edit

import "github.com/elves/elvish/parse-ng"

type TokenType int

const (
	ParserError TokenType = iota
	Bareword
	SingleQuoted
	DoubleQuoted
	Variable
	Sep
)

type Token struct {
	Type TokenType
	Text string
	Node parse.Node
}

var tokensBufferSize = 16

func tokenize(src string) ([]Token, error) {
	lastEnd := 0
	n, err := parse.Parse("[interactive code]", src)
	if n == nil {
		return []Token{{ParserError, src, nil}}, err
	}

	tokenCh := make(chan Token, tokensBufferSize)
	tokens := []Token{}
	tokensDone := make(chan bool)

	go func() {
		for token := range tokenCh {
			begin := token.Node.N().Begin
			if begin > lastEnd {
				tokens = append(tokens, Token{ParserError, src[lastEnd:begin], nil})
			}
			tokens = append(tokens, token)
			lastEnd = token.Node.N().End
		}
		tokensDone <- true
	}()
	produceTokens(n, tokenCh)
	close(tokenCh)

	<-tokensDone
	if lastEnd != len(src) {
		tokens = append(tokens, Token{ParserError, src[lastEnd:], nil})
	}
	return tokens, err
}

func produceTokens(n parse.Node, tokenCh chan<- Token) {
	if len(n.N().Children) == 0 {
		tokenType := ParserError
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
			}
		case *parse.Sep:
			tokenType = Sep
		}
		tokenCh <- Token{tokenType, n.N().SourceText, n}
	}
	for _, child := range n.N().Children {
		produceTokens(child, tokenCh)
	}
}
