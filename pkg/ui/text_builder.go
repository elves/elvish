package ui

import "strings"

// Methods of [TextBuilder] are fully exercised by other functions Concat, so
// there are no dedicated tests for it.

// TextBuilder can be used to efficiently build a [Text]. The zero value is
// ready to use. Do not copy a non-zero TextBuilder.
type TextBuilder struct {
	segs  []*Segment
	style Style
	text  strings.Builder
}

// WriteText appends t to the TextBuilder.
func (tb *TextBuilder) WriteText(t Text) {
	if len(t) == 0 {
		return
	}
	if tb.style == t[0].Style {
		// Merge the first segment of t with the pending segment.
		tb.text.WriteString(t[0].Text)
		t = t[1:]
		if len(t) == 0 {
			return
		}
	}
	// At this point, the first segment of t has a different style than the
	// pending segment (assuming that t is normal). Add the pending segment if
	// it's non-empty.
	if tb.text.Len() > 0 {
		tb.segs = append(tb.segs, &Segment{tb.style, tb.text.String()})
		tb.text.Reset()
	}
	// Add all segments from t except the last one.
	tb.segs = append(tb.segs, t[:len(t)-1]...)

	// Use the last segment of t as the pending segment.
	tb.style = t[len(t)-1].Style
	tb.text.WriteString(t[len(t)-1].Text)
}

// Text returns the [Text] that has been built.
func (tb *TextBuilder) Text() Text {
	if tb.Empty() {
		return nil
	}
	t := append(Text(nil), tb.segs...)
	return append(t, &Segment{tb.style, tb.text.String()})
}

// Empty returns nothing has been written to the TextBuilder yet.
func (tb *TextBuilder) Empty() bool {
	return len(tb.segs) == 0 && tb.text.Len() == 0
}

// Reset resets the TextBuilder to be empty.
func (tb *TextBuilder) Reset() {
	tb.segs = nil
	tb.style = Style{}
	tb.text.Reset()
}
