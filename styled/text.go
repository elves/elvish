package styled

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/elves/elvish/eval/vals"
)

// Text contains of a list of styled Segments.
type Text []Segment

func (t Text) Kind() string { return "styled-text" }

// Repr returns the representation of the current Text. It is just a wrapper
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

// Index provides access to the underlying Segments.
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

// Concat implements Text+string, Text+Segment and Text+Text.
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

// RConcat implements string+Text.
func (t Text) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		return Text(append([]Segment{{Text: lhs}}, t...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// Partition partitions the Text at n indicies into n+1 Text values.
func (t Text) Partition(indicies ...int) []Text {
	out := make([]Text, len(indicies)+1)
	segs := t
	consumedSegsLen := 0
	seg0Consumed := 0
	for i, idx := range indicies {
		text := make(Text, 0)
		for len(segs) > 0 && idx >= consumedSegsLen+len(segs[0].Text) {
			text = append(text, Segment{
				segs[0].Style, segs[0].Text[seg0Consumed:]})
			consumedSegsLen += len(segs[0].Text)
			seg0Consumed = 0
			segs = segs[1:]
		}
		if len(segs) > 0 && idx > consumedSegsLen {
			text = append(text, Segment{
				segs[0].Style, segs[0].Text[:idx-consumedSegsLen]})
			seg0Consumed = idx - consumedSegsLen
		}
		out[i] = text
	}
	trailing := make(Text, 0)
	for len(segs) > 0 {
		trailing = append(trailing, Segment{
			segs[0].Style, segs[0].Text[seg0Consumed:]})
		seg0Consumed = 0
		segs = segs[1:]
	}
	out[len(indicies)] = trailing
	return out
}
