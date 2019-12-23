package vals

import (
	"bytes"
	"strings"
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
		// Pretty-printing: Add a newline and indent+1 spaces.
		b.buf.WriteString("\n" + strings.Repeat(" ", b.indent+1))
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
		b.buf.WriteString("\n" + strings.Repeat(" ", b.indent))
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

// WritePair writes a new pair.
func (b *MapReprBuilder) WritePair(k string, indent int, v string) {
	if indent > 0 {
		b.inner.WriteElem("&" + k + "=\t" + v)
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
