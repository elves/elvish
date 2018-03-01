package vals

import (
	"bytes"
	"strings"

	"github.com/xiaq/persistent/vector"
)

// EmptyList is an empty list.
var EmptyList = vector.Empty

// MakeList creates a new List from values.
func MakeList(vs ...interface{}) vector.Vector {
	vec := vector.Empty
	for _, v := range vs {
		vec = vec.Cons(v)
	}
	return vec
}

// MakeList creates a new List from strings.
func MakeStringList(vs ...string) vector.Vector {
	vec := vector.Empty
	for _, v := range vs {
		vec = vec.Cons(v)
	}
	return vec
}

// ListReprBuilder helps to build Repr of list-like Values.
type ListReprBuilder struct {
	Indent int
	buf    bytes.Buffer
}

func (b *ListReprBuilder) WriteElem(v string) {
	if b.buf.Len() == 0 {
		b.buf.WriteByte('[')
	}
	if b.Indent >= 0 {
		// Pretty printing.
		//
		// Add a newline and indent+1 spaces, so that the
		// starting & lines up with the first pair.
		b.buf.WriteString("\n" + strings.Repeat(" ", b.Indent+1))
	} else if b.buf.Len() > 1 {
		b.buf.WriteByte(' ')
	}
	b.buf.WriteString(v)
}

func (b *ListReprBuilder) String() string {
	if b.buf.Len() == 0 {
		return "[]"
	}
	if b.Indent >= 0 {
		b.buf.WriteString("\n" + strings.Repeat(" ", b.Indent))
	}
	b.buf.WriteByte(']')
	return b.buf.String()
}
