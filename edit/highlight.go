package edit

import (
	"../parse"
	"../eval"
)

// Pseudo-ItemType's used by the Highlighter. They are given negative values
// to avoid conflict with genuine ItemType's. They can be regarded as some
// "sub-ItemType" of ItemBare, since only barewords are inspected and marked
// with these pseudo-ItemType's.
const (
	ItemValidCommand   parse.ItemType = -iota - 1
	ItemInvalidCommand
	ItemValidVariable
	ItemInvalidVariable
)

// A Highlighter is an enhanced Lexer with the knowledge of the validity of
// command names and variable names.
//
// Generally speaking, this is not possible without properly parsing and
// evaluating the source, performing its side effects. However in the case of
// single bareword terms, some heuristics can be applied. Be noted though, the
// heuristics can be quite fragile in case of syntax changes, so it's
// important to keep the syntax either simple (so that the heuristics is easy
// to implement) or stable (so that the heuristics doesn't have to be
// modified), or both.
type Highlighter struct {
	lexer *parse.Lexer
	ev *eval.Evaluator
	items chan parse.Item
}

func (hl *Highlighter) run() {
	command := true
	variable := false

	for token := range hl.lexer.Chan() {
		switch token.Typ {
		case parse.ItemBare:
			// Mangle token.Typ
			if command {
				// Check validity of command
				if _, _, err := hl.ev.ResolveCommand(token.Val); err == nil {
					token.Typ = ItemValidCommand
				} else {
					token.Typ = ItemInvalidCommand
				}
			} else if variable {
				if _, err := hl.ev.ResolveVar(token.Val); err == nil {
					token.Typ = ItemValidVariable
				} else {
					token.Typ = ItemInvalidVariable
				}
			}
			command = false
			variable = false
		case parse.ItemSemicolon, parse.ItemPipe, parse.ItemEndOfLine:
			// NOTE ItemPipe can also be the pipe in {|a b| command...}
			command = true
		case parse.ItemDollar:
			variable = true
		case parse.ItemSpace:
			variable = false
		default:
			command = false
			variable = false
		}
		// TODO highlight `echo` in `{ echo a }`
		hl.items <- token
	}
	close(hl.items)
}

func Highlight(name, input string, ev *eval.Evaluator) chan parse.Item {
	hl := &Highlighter{parse.Lex(name, input), ev, make(chan parse.Item)}
	go hl.run()
	return hl.items
}
