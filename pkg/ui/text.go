package ui

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/wcwidth"
)

// Text contains of a list of styled Segments.
type Text []*Segment

// T constructs a new Text with the given content and the given Styling's
// applied.
func T(s string, ts ...Styling) Text {
	return StyleText(Text{&Segment{Text: s}}, ts...)
}

// Kind returns "styled-text".
func (Text) Kind() string { return "ui:text" }

// Repr returns the representation of the current Text. It is just a wrapper
// around the containing Segments.
func (t Text) Repr(indent int) string {
	buf := new(bytes.Buffer)
	for _, s := range t {
		buf.WriteString(s.Repr(indent + 1))
	}
	return fmt.Sprintf("(ui:text %s)", buf.String())
}

// IterateKeys feeds the function with all valid indicies of the styled-text.
func (t Text) IterateKeys(fn func(interface{}) bool) {
	for i := 0; i < len(t); i++ {
		if !fn(strconv.Itoa(i)) {
			break
		}
	}
}

// Index provides access to the underlying styled-segment.
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
		return t.ConcatSegments(&Segment{Text: rhs}), nil
	case *Segment:
		return t.ConcatSegments(rhs), nil
	case Text:
		return t.ConcatSegments(rhs...), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// ConcatSegments returns a new Text with the new Text added to the end.
func (t Text) ConcatText(t2 Text) Text {
	return t.ConcatSegments(t2...)
}

// ConcatSegments returns a new Text with the new Segment's added to the end.
func (t Text) ConcatSegments(segs ...*Segment) Text {
	return Text(append(append(Text(nil), t...), segs...))
}

// RConcat implements string+Text.
func (t Text) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		return Text(append([]*Segment{{Text: lhs}}, t...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// Partition partitions the Text at n indicies into n+1 Text values.
func (t Text) Partition(indicies ...int) []Text {
	out := make([]Text, len(indicies)+1)
	segs := t.Clone()
	for i, idx := range indicies {
		toConsume := idx
		if i > 0 {
			toConsume -= indicies[i-1]
		}
		for len(segs) > 0 && toConsume > 0 {
			if len(segs[0].Text) <= toConsume {
				out[i] = append(out[i], segs[0])
				toConsume -= len(segs[0].Text)
				segs = segs[1:]
			} else {
				out[i] = append(out[i], &Segment{segs[0].Style, segs[0].Text[:toConsume]})
				segs[0] = &Segment{segs[0].Style, segs[0].Text[toConsume:]}
				toConsume = 0
			}
		}
	}
	if len(segs) > 0 {
		// Don't use segs directly to avoid memory leak
		out[len(indicies)] = append(Text(nil), segs...)
	}
	return out
}

// Clone returns a deep copy of Text.
func (t Text) Clone() Text {
	newt := make(Text, len(t))
	for i, seg := range t {
		newt[i] = seg.Clone()
	}
	return newt
}

// CountRune counts the number of times a rune occurs in a Text.
func (t Text) CountRune(r rune) int {
	n := 0
	for _, seg := range t {
		n += seg.CountRune(r)
	}
	return n
}

// CountLines counts the number of lines in a Text. It is equal to
// t.CountRune('\n') + 1.
func (t Text) CountLines() int {
	return t.CountRune('\n') + 1
}

// SplitByRune splits a Text by the given rune.
func (t Text) SplitByRune(r rune) []Text {
	// Call SplitByRune for each constituent Segment, and "paste" the pairs of
	// subsegments across the segment border. For instance, if Text has 3
	// Segments a, b, c that results in a1, a2, a3, b1, b2, c1, then a3 and b1
	// as well as b2 and c1 are pasted together, and the return value is [a1],
	// [a2], [a3, b1], [b2, c1]. Pasting can happen coalesce: for instance, if
	// Text has 3 Segments a, b, c that results in a1, a2, b1, c1, the return
	// value will be [a1], [a2, b1, c1].
	var result []Text
	var paste Text
	for _, seg := range t {
		subSegs := seg.SplitByRune(r)
		if len(subSegs) == 1 {
			// Only one subsegment. Just paste.
			paste = append(paste, subSegs[0])
			continue
		}
		// Paste the previous trailing segments with the first subsegment, and
		// add it as a Text.
		result = append(result, append(paste, subSegs[0]))
		// For the subsegments in the middle, just add then as is.
		for i := 1; i < len(subSegs)-1; i++ {
			result = append(result, Text{subSegs[i]})
		}
		// The last segment becomes the new paste.
		paste = Text{subSegs[len(subSegs)-1]}
	}
	if len(paste) > 0 {
		result = append(result, paste)
	}
	return result
}

// TrimWcwidth returns the largest prefix of t that does not exceed the given
// visual width.
func (t Text) TrimWcwidth(wmax int) Text {
	var newt Text
	for _, seg := range t {
		w := wcwidth.Of(seg.Text)
		if w >= wmax {
			newt = append(newt,
				&Segment{seg.Style, wcwidth.Trim(seg.Text, wmax)})
			break
		}
		wmax -= w
		newt = append(newt, seg)
	}
	return newt
}

// String returns a string representation of the styled text. This now always
// assumes VT-style terminal output.
//
// TODO: Make string conversion sensible to environment, e.g. use HTML when
// output is web.
func (t Text) String() string {
	return t.VTString()
}

// VTString renders the styled text using VT-style escape sequences.
func (t Text) VTString() string {
	var buf bytes.Buffer
	for _, seg := range t {
		buf.WriteString(seg.VTString())
	}
	return buf.String()
}
