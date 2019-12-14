package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

// Segment is a string that has some style applied to it.
type Segment struct {
	Style
	Text string
}

// Kind returns "styled-segment".
func (*Segment) Kind() string { return "styled-segment" }

// Repr returns the representation of this Segment. The string can be used to
// construct an identical Segment. Unset or default attributes are skipped. If
// the Segment represents an unstyled string only this string is returned.
func (s *Segment) Repr(indent int) string {
	buf := new(bytes.Buffer)
	addIfNotEqual := func(key string, val, cmp interface{}) {
		if val != cmp {
			fmt.Fprintf(buf, "&%s=%s ", key, vals.Repr(val, 0))
		}
	}

	addIfNotEqual("fg-color", s.Foreground, "")
	addIfNotEqual("bg-color", s.Background, "")
	addIfNotEqual("bold", s.Bold, false)
	addIfNotEqual("dim", s.Dim, false)
	addIfNotEqual("italic", s.Italic, false)
	addIfNotEqual("underlined", s.Underlined, false)
	addIfNotEqual("blink", s.Blink, false)
	addIfNotEqual("inverse", s.Inverse, false)

	if buf.Len() == 0 {
		return s.Text
	}

	return fmt.Sprintf("(styled-segment %s %s)", s.Text, strings.TrimSpace(buf.String()))
}

// IterateKeys feeds the function with all valid attributes of styled-segment.
func (*Segment) IterateKeys(fn func(v interface{}) bool) {
	util.Feed(fn, "text", "fg-color", "bg-color", "bold", "dim", "italic", "underlined", "blink", "inverse")
}

// Index provides access to the attributes of a styled-segment.
func (s *Segment) Index(k interface{}) (v interface{}, ok bool) {
	switch k {
	case "text":
		v = s.Text
	case "fg-color":
		v = s.Foreground
	case "bg-color":
		v = s.Background
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

	if v == "" {
		v = "default"
	}

	return v, v != nil
}

// Concat implements Segment+string, Segment+Segment and Segment+Text.
func (s *Segment) Concat(v interface{}) (interface{}, error) {
	switch rhs := v.(type) {
	case string:
		return Text{
			s,
			&Segment{Text: rhs},
		}, nil
	case *Segment:
		return Text{s, rhs}, nil
	case Text:
		return Text(append([]*Segment{s}, rhs...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// RConcat implements string+Segment.
func (s *Segment) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		return Text{
			&Segment{Text: lhs},
			s,
		}, nil
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
