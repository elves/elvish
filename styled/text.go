package styled

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

// Text contains of a list of styled Segments.
type Text []*Segment

// Plain returns an unstyled Text.
func Plain(s string) Text {
	return Text{PlainSegment(s)}
}

// MakeText makes a text by taking a string and applying the given transformers.
func MakeText(s string, transformers ...string) Text {
	t := Plain(s)
	for _, transformer := range transformers {
		t = Transform(t, transformer)
	}
	return t
}

// Kind returns "styled-text".
func (Text) Kind() string { return "styled-text" }

// Repr returns the representation of the current Text. It is just a wrapper
// around the containing Segments.
func (t Text) Repr(indent int) string {
	buf := new(bytes.Buffer)
	for _, s := range t {
		buf.WriteString(s.Repr(indent + 1))
	}
	return fmt.Sprintf("(styled %s)", buf.String())
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
		return Text(append(t, &Segment{Text: rhs})), nil
	case *Segment:
		return Text(append(t, rhs)), nil
	case Text:
		return Text(append(t, rhs...)), nil
	}

	return nil, vals.ErrConcatNotImplemented
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
	segs := t
	consumedSegsLen := 0
	seg0Consumed := 0
	for i, idx := range indicies {
		text := make(Text, 0)
		for len(segs) > 0 && idx >= consumedSegsLen+len(segs[0].Text) {
			text = append(text, &Segment{
				segs[0].Style, segs[0].Text[seg0Consumed:]})
			consumedSegsLen += len(segs[0].Text)
			seg0Consumed = 0
			segs = segs[1:]
		}
		if len(segs) > 0 && idx > consumedSegsLen {
			text = append(text, &Segment{
				segs[0].Style, segs[0].Text[:idx-consumedSegsLen]})
			seg0Consumed = idx - consumedSegsLen
		}
		out[i] = text
	}
	trailing := make(Text, 0)
	for len(segs) > 0 {
		trailing = append(trailing, &Segment{
			segs[0].Style, segs[0].Text[seg0Consumed:]})
		seg0Consumed = 0
		segs = segs[1:]
	}
	out[len(indicies)] = trailing
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
		w := util.Wcswidth(seg.Text)
		if w >= wmax {
			newt = append(newt,
				&Segment{seg.Style, util.TrimWcwidth(seg.Text, wmax)})
			break
		}
		wmax -= w
		newt = append(newt, seg)
	}
	return newt
}
