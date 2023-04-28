package vals

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// ListReprBuilder helps to build Repr of list-like Values.
type ListReprBuilder struct {
	indent int
	buf    bytes.Buffer
}

// NewListReprBuilder makes a new ListReprBuilder.
func NewListReprBuilder(indent int) *ListReprBuilder {
	return &ListReprBuilder{indent: indent}
}

// WriteElem writes a new element.
func (b *ListReprBuilder) WriteElem(v string) {
	if b.buf.Len() == 0 {
		b.buf.WriteByte('[')
	}
	if b.indent >= 0 {
		// Pretty-printing: Add a newline and indent the list values.
		b.buf.WriteByte('\n')
		b.buf.WriteString(strings.Repeat(" ", 2*(b.indent+1)))
	} else if b.buf.Len() > 1 {
		b.buf.WriteByte(' ')
	}
	b.buf.WriteString(v)
}

// String returns the representation that has been built. After it is called,
// the ListReprBuilder may no longer be used.
func (b *ListReprBuilder) String() string {
	if b.buf.Len() == 0 {
		return "[]"
	}
	if b.indent >= 0 {
		b.buf.WriteByte('\n')
		b.buf.WriteString(strings.Repeat(" ", 2*b.indent))
	}
	b.buf.WriteByte(']')
	return b.buf.String()
}

// MapReprBuilder helps building the Repr of a Map. It is also useful for
// implementing other Map-like values. The zero value of a MapReprBuilder is
// ready to use.
type MapReprBuilder struct {
	inner ListReprBuilder
}

// NewMapReprBuilder makes a new MapReprBuilder.
func NewMapReprBuilder(indent int) *MapReprBuilder {
	return &MapReprBuilder{ListReprBuilder{indent: indent}}
}

// WritePair writes a new key:value pair. The caller should calculate the
// maximum displayable width of any key in the set of keys and pass that value
// to this method in order to ensure the values are vertically aligned.
func (b *MapReprBuilder) WritePair(maxKeyWidth int, k string, indent int, v string) {
	if indent > 0 {
		padding := 1 + maxKeyWidth - runewidth.StringWidth(k)
		b.inner.WriteElem("&" + k + fmt.Sprintf("%-*s", padding, "=") + v)
	} else {
		b.inner.WriteElem("&" + k + "=" + v)
	}
}

// String returns the representation that has been built. After it is called,
// the MapReprBuilder should no longer be used.
func (b *MapReprBuilder) String() string {
	s := b.inner.String()
	if s == "[]" {
		return "[&]"
	}
	return s
}
