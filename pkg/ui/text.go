package ui

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/wcwidth"
)

// Text contains of a list of styled Segments.
//
// When only functions in this package are used to manipulate Text instances,
// they will always satisfy the following properties:
//
//   - If the Text is empty, it is nil (not a non-nil slice of size 0).
//
//   - No [Segment] in it has an empty Text field.
//
//   - No two adjacent [Segment] instances have the same [Style].
type Text []*Segment

// T constructs a new Text with the given content and the given Styling's
// applied.
func T(s string, ts ...Styling) Text {
	if s == "" {
		return nil
	}
	return StyleText(Text{&Segment{Text: s}}, ts...)
}

// Concat concatenates multiple Text's into one.
func Concat(texts ...Text) Text {
	var tb TextBuilder
	for _, text := range texts {
		tb.WriteText(text)
	}
	return tb.Text()
}

// Kind returns "styled-text".
func (Text) Kind() string { return "ui:text" }

// Repr returns the representation of the current Text. It is just a wrapper
// around the containing Segments.
func (t Text) Repr(indent int) string {
	buf := new(bytes.Buffer)
	for _, s := range t {
		buf.WriteByte(' ')
		buf.WriteString(s.Repr(indent + 1))
	}
	return fmt.Sprintf("[^styled%s]", buf.String())
}

// IterateKeys feeds the function with all valid indices of the styled-text.
func (t Text) IterateKeys(fn func(any) bool) {
	for i := 0; i < len(t); i++ {
		if !fn(strconv.Itoa(i)) {
			break
		}
	}
}

// Index provides access to the underlying styled-segment.
func (t Text) Index(k any) (any, error) {
	index, err := vals.ConvertListIndex(k, len(t))
	if err != nil {
		return nil, err
	} else if index.Slice {
		return t[index.Lower:index.Upper], nil
	} else {
		return t[index.Lower], nil
	}
}

// Concat implements Text+string, Text+number, Text+Segment and Text+Text.
func (t Text) Concat(rhs any) (any, error) {
	switch rhs := rhs.(type) {
	case string:
		return Concat(t, T(rhs)), nil
	case int, *big.Int, *big.Rat, float64:
		return Concat(t, T(vals.ToString(rhs))), nil
	case *Segment:
		return Concat(t, Text{rhs}), nil
	case Text:
		return Concat(t, rhs), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// RConcat implements string+Text and number+Text.
func (t Text) RConcat(lhs any) (any, error) {
	switch lhs := lhs.(type) {
	case string:
		return Concat(T(lhs), t), nil
	case int, *big.Int, *big.Rat, float64:
		return Concat(T(vals.ToString(lhs)), t), nil
	}

	return nil, vals.ErrConcatNotImplemented
}

// Partition partitions the Text at n indices into n+1 Text values.
func (t Text) Partition(indices ...int) []Text {
	out := make([]Text, len(indices)+1)
	segs := t.Clone()
	for i, idx := range indices {
		toConsume := idx
		if i > 0 {
			toConsume -= indices[i-1]
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
		out[len(indices)] = append(Text(nil), segs...)
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
	if len(t) == 0 {
		return nil
	}
	// Call SplitByRune for each constituent Segment, and "paste" the pairs of
	// subsegments across the segment border. For instance, if Text has 3
	// Segments a, b, c that results in a1, a2, a3, b1, b2, c1, then a3 and b1
	// as well as b2 and c1 are pasted together, and the return value is [a1],
	// [a2], [a3, b1], [b2, c1]. Pasting can coalesce: for instance, if
	// Text has 3 Segments a, b, c that results in a1, a2, b1, c1, the return
	// value will be [a1], [a2, b1, c1].
	var result []Text
	var paste TextBuilder
	for _, seg := range t {
		subSegs := seg.SplitByRune(r)
		// Paste the first segment.
		paste.WriteText(TextFromSegment(subSegs[0]))
		if len(subSegs) == 1 {
			// Only one subsegment. Keep the paste active.
			continue
		}
		// Add the paste and reset it.
		result = append(result, paste.Text())
		paste.Reset()
		// For the subsegments in the middle, just add then as is.
		for i := 1; i < len(subSegs)-1; i++ {
			result = append(result, TextFromSegment(subSegs[i]))
		}
		// The last segment becomes the new paste.
		paste.WriteText(TextFromSegment(subSegs[len(subSegs)-1]))
	}
	result = append(result, paste.Text())
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

// VTString renders the styled text using VT-style escape sequences. Any
// existing SGR state will be cleared.
func (t Text) VTString() string {
	var sb strings.Builder
	clean := false
	for _, seg := range t {
		sgr := seg.SGR()
		if sgr == "" {
			if !clean {
				sb.WriteString("\033[m")
			}
			clean = true
		} else {
			if clean {
				sb.WriteString("\033[" + sgr + "m")
			} else {
				sb.WriteString("\033[;" + sgr + "m")
			}
			clean = false
		}
		sb.WriteString(seg.Text)
	}
	if !clean {
		sb.WriteString("\033[m")
	}
	return sb.String()
}

// TextFromSegment returns a [Text] with just seg if seg.Text is non-empty.
// Otherwise it returns nil.
func TextFromSegment(seg *Segment) Text {
	if seg.Text == "" {
		return nil
	}
	return Text{seg}
}
