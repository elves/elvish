// Derived from stdlib package text/template/parse.

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Item represents a token or text string returned from the scanner.
type Item struct {
	Typ ItemType // The type of this Item.
	Pos Pos      // The starting position, in bytes, of this Item in the input string.
	Val string   // The value of this Item.
	End ItemEnd  // How an Item ends.
}

func (i Item) String() string {
	if i.Typ == ItemError {
		return i.Val
	}
	return fmt.Sprintf("%s %q", ItemTypeNames[i.Typ], i.Val)
}

// ItemType identifies the type of lex items.
type ItemType int

const (
	ItemError        ItemType = iota // error occurred; value is text of error
	ItemEOF
	ItemEndOfLine    // a single EOL
	ItemSpace        // run of spaces separating arguments
	ItemBare         // a bare string literal
	ItemSingleQuoted // a single-quoted string literal
	ItemDoubleQuoted // a double-quoted string literal
	ItemRedirLeader  // IO redirection leader
	ItemPipe         // pipeline connector, '|'
	ItemLParen       // left paren '('
	ItemRParen       // right paren ')'
)

var ItemTypeNames []string = []string {
	"ItemError",
	"ItemEOF",
	"ItemEndOfLine",
	"ItemSpace",
	"ItemBare",
	"ItemSingleQuoted",
	"ItemDoubleQuoted",
	"ItemRedirLeader",
	"ItemPipe",
	"ItemLParen",
	"ItemRParen",
}

// ItemEnd describes the ending of lex items.
type ItemEnd int

const (
	MayTerminate ItemEnd = 1 << iota
	MayContinue
	ItemTerminated   ItemEnd = MayTerminate
	ItemUnterminated ItemEnd = MayContinue
	ItemAmbiguious   ItemEnd = MayTerminate | MayContinue
)

const Eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Lexer) stateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this Item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent Item returned by NextItem
	items      chan Item // channel of scanned items
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return Eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// emit passes an Item back to the client.
func (l *Lexer) emit(t ItemType, e ItemEnd) {
	l.items <- Item{t, l.start, l.input[l.start:l.pos], e}
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous Item returned by NextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *Lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.NextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{ItemError, l.start, fmt.Sprintf(format, args...), ItemEnd(0)}
	return nil
}

// NextItem returns the next Item from the input.
func (l *Lexer) NextItem() Item {
	item := <-l.items
	l.lastPos = item.Pos
	return item
}

// Chan returns a channel of Item's.
func (l *Lexer) Chan() chan Item {
	return l.items
}

// Lex creates a new scanner for the input string.
func Lex(name, input string) *Lexer {
	l := &Lexer{
		name:       name,
		input:      input,
		items:      make(chan Item),
	}
	go l.run()
	return l
}

// run runs the state machine for the Lexer.
func (l *Lexer) run() {
	for l.state = lexAny; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

// state functions

// lexAny is the default state.
func lexAny(l *Lexer) stateFn {
	switch r := l.next(); {
	case r == Eof:
		l.emit(ItemEOF, ItemTerminated)
		return nil
	case isSpace(r):
		return lexSpace
	case r == '>' || r == '<':
		l.backup()
		return lexRedirLeader
	case r == '\n':
		return lexEndOfLine
	case r == '\'':
		return lexSingleQuoted
	case r == '"':
		return lexDoubleQuoted
	case r == '|':
		l.emit(ItemPipe, ItemTerminated)
		return lexAny
	case r == '(':
		l.emit(ItemLParen, ItemTerminated)
		return lexAny
	case r == ')':
		l.emit(ItemRParen, ItemTerminated)
		return lexAny
	default:
		return lexBare
	}
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *Lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(ItemSpace, ItemAmbiguious)
	return lexAny
}

// lexRedirLeader scans an IO redirection leader.
// It is started by one of < <> > >> and may be followed immediately by a
// string surrounded by square brackets. The internal structure of the string
// is not checked here.
func lexRedirLeader(l *Lexer) stateFn {
	switch l.next() {
	case '<', '>':
		if l.peek() == '>' {
			l.next()
		}
	default:
		panic("unreachable")
	}

	if l.peek() == '[' {
loop:
		for {
			switch l.next() {
			case ']':
				l.emit(ItemRedirLeader, ItemTerminated)
				break loop
			case Eof:
				l.emit(ItemRedirLeader, ItemUnterminated)
				break loop
			}
		}
	} else {
		l.emit(ItemRedirLeader, ItemAmbiguious)
	}

	return lexAny
}

// lexEndOfLine scans a single EOL, which has already been seen.
func lexEndOfLine(l *Lexer) stateFn {
	l.emit(ItemEndOfLine, ItemTerminated)
	return lexAny
}

// lexBare scans a bare string.
// The first rune has already been seen.
func lexBare(l *Lexer) stateFn {
	for !terminatesBare(l.peek()) {
		l.next()
	}
	l.emit(ItemBare, ItemAmbiguious)
	return lexAny
}

func terminatesBare(r rune) bool {
	return isSpace(r) || r == '\n' || r == '(' || r == ')' || r == Eof
}

// lexSingleQuoted scans a single-quoted string.
// The opening quote has already been seen.
func lexSingleQuoted(l *Lexer) stateFn {
	const quote = '\''
loop:
	for {
		switch l.next() {
		case Eof, '\n':
			l.emit(ItemSingleQuoted, ItemUnterminated)
			return lexAny
		case quote:
			if l.peek() != quote {
				break loop
			}
			l.next()
		}
	}
	l.emit(ItemSingleQuoted, ItemAmbiguious)
	return lexAny
}

// lexDoubleQuoted scans a double-quoted string.
// The opening quote has already been seen.
func lexDoubleQuoted(l *Lexer) stateFn {
loop:
	for {
		switch l.next() {
		case '\\':
			if r:= l.next(); r != Eof && r != '\n' {
				break
			}
			fallthrough
		case Eof, '\n':
			l.emit(ItemDoubleQuoted, ItemUnterminated)
			return lexAny
		case '"':
			break loop
		}
	}
	l.emit(ItemDoubleQuoted, ItemTerminated)
	return lexAny
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
