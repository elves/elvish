package edit

import (
	"github.com/xiaq/elvish/eval"
	"github.com/xiaq/elvish/parse"
)

// Pseudo-ItemType's used by the Highlighter. They are given negative values
// to avoid conflict with genuine ItemType's. They can be regarded as some
// "sub-ItemType" of ItemBare, since only barewords are inspected and marked
// with these pseudo-ItemType's.
const (
	ItemValidCommand parse.ItemType = -iota - 1
	ItemInvalidCommand
	ItemValidVariable
	ItemInvalidVariable
)

// Highlighter is an enhanced Lexer with the knowledge of the validity of
// command names and variable names.
//
// Generally speaking, this is not possible without properly parsing and
// evaluating the source, performing its side effects. However in the case of
// single bareword expressions, some heuristics can be applied. Be noted
// though, the heuristics can be quite fragile in case of syntax changes, so
// it's important to keep the syntax either simple (so that the heuristics is
// easy to implement) or stable (so that the heuristics doesn't have to be
// modified), or both.
type Highlighter struct {
	lexer *parse.Lexer
	ev    *eval.Evaluator
	items chan parse.Item
}

func (hl *Highlighter) variable(token parse.Item) {
	if token.Typ == parse.ItemBare {
		// XXX Disabled until Compiler implements it
		if true {
			// if _, err := hl.ev.ResolveVar(token.Val); err == nil {
			token.Typ = ItemValidVariable
		} else {
			token.Typ = ItemInvalidVariable
		}
	}
	hl.items <- token
}

func (hl *Highlighter) command(token parse.Item) {
	if token.Typ == parse.ItemSpace {
		hl.items <- token
		token = <-hl.lexer.Chan()
	}
	if token.Typ == parse.ItemBare {
		// Check validity of command
		// XXX Disabled until Compiler implements it
		if true {
			// if _, _, err := hl.ev.ResolveCommand(token.Val); err == nil {
			token.Typ = ItemValidCommand
		} else {
			token.Typ = ItemInvalidCommand
		}
	}
	hl.items <- token
}

func (hl *Highlighter) run() {
	tokens := hl.lexer.Chan()

	// First token is command
	// TODO Support other more interesting commands as soon as checker allows
	hl.command(<-tokens)
Loop:
	for {
		token := <-tokens
		switch token.Typ {
		case parse.ItemDollar:
			hl.items <- token
			hl.variable(<-tokens)
		case parse.ItemSemicolon, parse.ItemPipe, parse.ItemEndOfLine,
			parse.ItemLParen, parse.ItemQuestionLParen:
			hl.items <- token
			hl.command(<-tokens)
		case parse.ItemLBrace:
			hl.items <- token
			token = <-tokens
			switch token.Typ {
			case parse.ItemPipe:
				hl.items <- token
			Args:
				for {
					token = <-tokens
					hl.items <- token
					switch token.Typ {
					case parse.ItemPipe:
						break Args
					case parse.ItemError, parse.ItemEOF:
						break Loop
					}
				}
				hl.command(<-tokens)
			case parse.ItemSpace:
				hl.command(token)
			default:
				hl.items <- token
			}
		case parse.ItemError, parse.ItemEOF:
			hl.items <- token
			break Loop
		default:
			hl.items <- token
		}
	}

	close(hl.items)
}

func Highlight(name, input string, ev *eval.Evaluator) chan parse.Item {
	hl := &Highlighter{parse.Lex(name, input), ev, make(chan parse.Item)}
	go hl.run()
	return hl.items
}
