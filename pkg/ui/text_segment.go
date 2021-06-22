package ui

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"src.elv.sh/pkg/eval/vals"
)

// Segment is a string that has some style applied to it.
type Segment struct {
	Style
	Text string
}

// Kind returns "styled-segment".
func (*Segment) Kind() string { return "ui:text-segment" }

// Repr returns the representation of this Segment. The string can be used to
// construct an identical Segment. Unset or default attributes are skipped. If
// the Segment represents an unstyled string only this string is returned.
func (s *Segment) Repr(int) string {
	buf := new(bytes.Buffer)
	addIfNotEqual := func(key string, val, cmp interface{}) {
		if val != cmp {
			var valString string
			if c, ok := val.(Color); ok {
				valString = c.String()
			} else {
				valString = vals.Repr(val, 0)
			}
			fmt.Fprintf(buf, "&%s=%s ", key, valString)
		}
	}

	addIfNotEqual("fg-color", s.Foreground, nil)
	addIfNotEqual("bg-color", s.Background, nil)
	addIfNotEqual("bold", s.Bold, false)
	addIfNotEqual("dim", s.Dim, false)
	addIfNotEqual("italic", s.Italic, false)
	addIfNotEqual("underlined", s.Underlined, false)
	addIfNotEqual("blink", s.Blink, false)
	addIfNotEqual("inverse", s.Inverse, false)

	if buf.Len() == 0 {
		return s.Text
	}

	return fmt.Sprintf("(ui:text-segment %s %s)", s.Text, strings.TrimSpace(buf.String()))
}

// IterateKeys feeds the function with all valid attributes of styled-segment.
func (*Segment) IterateKeys(fn func(v interface{}) bool) {
	vals.Feed(fn, "text", "fg-color", "bg-color", "bold", "dim", "italic", "underlined", "blink", "inverse")
}

// Index provides access to the attributes of a styled-segment.
func (s *Segment) Index(k interface{}) (v interface{}, ok bool) {
	switch k {
	case "text":
		v = s.Text
	case "fg-color":
		if s.Foreground == nil {
			return "default", true
		}
		return s.Foreground.String(), true
	case "bg-color":
		if s.Background == nil {
			return "default", true
		}
		return s.Background.String(), true
	case "bold":
		v = s.Bold
	case "dim":
		v = s.Dim
	case "italic":
		v = s.Italic
	case "underlined":
		v = s.Underlined
	case "blink":
		v = s.Blink
	case "inverse":
		v = s.Inverse
	}

	return v, v != nil
}

// Concat implements Segment+string, Segment+float64, Segment+Segment and
// Segment+Text.
func (s *Segment) Concat(v interface{}) (interface{}, error) {
	switch rhs := v.(type) {
	case string:
		return Text{s, &Segment{Text: rhs}}, nil
	case *Segment:
		return Text{s, rhs}, nil
	case Text:
		return Text(append([]*Segment{s}, rhs...)), nil
	case int, *big.Int, *big.Rat, float64:
		return Text{s, &Segment{Text: vals.ToString(rhs)}}, nil
	}
	return nil, vals.ErrConcatNotImplemented
}

// RConcat implements string+Segment and float64+Segment.
func (s *Segment) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		return Text{&Segment{Text: lhs}, s}, nil
	case int, *big.Int, *big.Rat, float64:
		return Text{&Segment{Text: vals.ToString(lhs)}, s}, nil
	}
	return nil, vals.ErrConcatNotImplemented
}

// Clone returns a copy of the Segment.
func (s *Segment) Clone() *Segment {
	value := *s
	return &value
}

// CountRune counts the number of times a rune occurs in a Segment.
func (s *Segment) CountRune(r rune) int {
	return strings.Count(s.Text, string(r))
}

// SplitByRune splits a Segment by the given rune.
func (s *Segment) SplitByRune(r rune) []*Segment {
	splitTexts := strings.Split(s.Text, string(r))
	splitSegs := make([]*Segment, len(splitTexts))
	for i, splitText := range splitTexts {
		splitSegs[i] = &Segment{s.Style, splitText}
	}
	return splitSegs
}

// String returns a string representation of the styled segment. This now always
// assumes VT-style terminal output.
// TODO: Make string conversion sensible to environment, e.g. use HTML when
// output is web.
func (s *Segment) String() string {
	return s.VTString()
}

// VTString renders the styled segment using VT-style escape sequences.
func (s *Segment) VTString() string {
	sgr := s.SGR()
	if sgr == "" {
		return s.Text
	}
	return fmt.Sprintf("\033[%sm%s\033[m", sgr, s.Text)
}
