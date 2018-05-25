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
	add := func(key string, val interface{}) {
		fmt.Fprintf(buf, "&%s=%s ", key, vals.Repr(val, 0))
	}

	for k, v := range s.Style.ToMap() {
		switch v := v.(type) {
		case string:
			// todo: Display default color?
			if v != "default" {
				add(k, v)
			}
		case bool:
			if v {
				add(k, v)
			}
		}
	}

	if buf.Len() == 0 {
		return s.Text
	}

	return fmt.Sprintf("(styled-segment %s %s)", s.Text, strings.TrimSpace(buf.String()))
}

func (s Segment) IterateKeys(fn func(v interface{}) bool) {
	for k := range s.Style.ToMap() {
		if !fn(k) {
			break
		}
	}
}

// Provides access to the attributes of the Segment.
func (s Segment) Index(k interface{}) (v interface{}, ok bool) {
	if k == "text" {
		return s.Text, true
	} else if k, ok := k.(string); ok {
		m := s.Style.ToMap()
		if v, ok := m[k]; ok {
			return v, true
		}
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
