package styled

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/elves/elvish/eval/vals"
)

// A styled Segment is a string that has some style applied to it.
type Segment struct {
	Style
	Text string
}

func (Segment) Kind() string { return "styled-segment" }

// Returns the representation of this Segment. The string can be used to construct
// an identical Segment. Unset or default attributes are skipped. If the Segment
// represents an unstyled string only this string is returned.
func (s Segment) Repr(indent int) string {
	buf := new(bytes.Buffer)
	addIfNotEqual := func(key string, val, cmp interface{}) {
		if val != cmp {
			fmt.Fprintf(buf, "&%s=%s ", key, vals.Repr(val, 0))
		}
	}

	addIfNotEqual("fg-color", s.Foreground, "default")
	addIfNotEqual("bg-color", s.Background, "default")
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

func (s Segment) IterateKeys(fn func(v interface{}) bool) {
	vals.Feed(fn, "text", "fg-color", "bg-color", "bold", "dim", "italic", "underlined", "blink", "inverse")
}

// Provides access to the attributes of the Segment.
func (s Segment) Index(k interface{}) (v interface{}, ok bool) {
	switch k {
	case "text":
		return s.Text, true
	case "fg-color":
		return s.Foreground, true
	case "bg-color":
		return s.Background, true
	case "bold":
		return s.Bold, true
	case "dim":
		return s.Dim, true
	case "italic":
		return s.Italic, true
	case "underlined":
		return s.Underlined, true
	case "blink":
		return s.Blink, true
	case "inverse":
		return s.Inverse, true
	}

	return nil, false
}

// Implements Segment+string, Segment+Segment and Segment+Text.
func (s Segment) Concat(v interface{}) (interface{}, error) {
	switch rhs := v.(type) {
	case string:
		return Text{
			s,
			Segment{Text: rhs},
		}, nil
	case *Segment:
		return Text{s, *rhs}, nil
	case *Text:
		return Text(append([]Segment{s}, *rhs...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// Implements string+Segment.
func (s Segment) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		return Text{
			Segment{Text: lhs},
			s,
		}, nil
	}

	return nil, vals.ErrConcatNotImplemented
}
