package glob

import (
	"bytes"
	"unicode/utf8"
)

// Parse parses a pattern.
func Parse(s string) Pattern {
	segments := []Segment{}
	add := func(seg Segment) {
		segments = append(segments, seg)
	}
	p := &parser{s, 0, 0}

rune:
	for {
		r := p.next()
		switch r {
		case eof:
			break rune
		case '?':
			add(Wild{Question, false, nil})
		case '*':
			n := 1
			for p.next() == '*' {
				n++
			}
			p.backup()
			if n == 1 {
				add(Wild{Star, false, nil})
			} else {
				add(Wild{StarStar, false, nil})
			}
		case '/':
			for p.next() == '/' {
			}
			p.backup()
			add(Slash{})
		default:
			var literal bytes.Buffer
		literal:
			for {
				switch r {
				case '?', '*', '/', eof:
					break literal
				case '\\':
					r = p.next()
					if r == eof {
						break literal
					}
					literal.WriteRune(r)
				default:
					literal.WriteRune(r)
				}
				r = p.next()
			}
			p.backup()
			add(Literal{literal.String()})
		}
	}
	return Pattern{segments, ""}
}

// TODO(xiaq): Contains duplicate code with parse/parser.go.

type parser struct {
	src     string
	pos     int
	overEOF int
}

const eof rune = -1

func (ps *parser) next() rune {
	if ps.pos == len(ps.src) {
		ps.overEOF++
		return eof
	}
	r, s := utf8.DecodeRuneInString(ps.src[ps.pos:])
	ps.pos += s
	return r
}

func (ps *parser) backup() {
	if ps.overEOF > 0 {
		ps.overEOF--
		return
	}
	_, s := utf8.DecodeLastRuneInString(ps.src[:ps.pos])
	ps.pos -= s
}
