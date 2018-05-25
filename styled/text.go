package styled

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/elves/elvish/eval/vals"
)

// A styled Text contains of a list of styled Segments.
type Text []Segment

func (t Text) Kind() string { return "styled-text" }

// Returns the representation of the current Text. It is just a wrapper
// around the containing Segments.
func (t Text) Repr(indent int) string {
	buf := new(bytes.Buffer)
	for _, s := range t {
		buf.WriteString(s.Repr(indent + 1))
	}
	return fmt.Sprintf("(styled %s)", buf.String())
}

func (t Text) IterateKeys(fn func(interface{}) bool) {
	for i := 0; i < len(t); i++ {
		if !fn(strconv.Itoa(i)) {
			break
		}
	}
}

// Provides access to the underlying Segments.
func (t Text) Index(k interface{}) (interface{}, error) {
	index, err := vals.ConvertListIndex(k, len(t))
	if err != nil {
		return nil, err
	} else if index.Slice {
		return t[index.Lower:index.Upper], nil
	} else {
		return t[index.Lower], nil
	}
}

// Implements Text+string, Text+Segment and Text+Text.
func (t Text) Concat(v interface{}) (interface{}, error) {
	switch rhs := v.(type) {
	case string:
		return Text(append(t, Segment{Text: rhs})), nil
	case *Segment:
		return Text(append(t, *rhs)), nil
	case *Text:
		return Text(append(t, *rhs...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// Implements string+Text.
func (t Text) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		return Text(append([]Segment{{Text: lhs}}, t...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
}
