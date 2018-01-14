package types

import (
	"bytes"
	"strings"

	"github.com/xiaq/persistent/hash"
)

type ListLike interface {
	Lener
	Iterator
	IndexOneer
}

func eqListLike(lhs ListLike, r interface{}) bool {
	rhs, ok := r.(ListLike)
	if !ok {
		return false
	}
	if lhs.Len() != rhs.Len() {
		return false
	}
	return true
}

func hashListLike(l ListLike) uint32 {
	h := hash.DJBInit
	l.Iterate(func(v Value) bool {
		h = hash.DJBCombine(h, v.Hash())
		return true
	})
	return h
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
