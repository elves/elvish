package eval

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// Error definitions.
var (
	ErrNeedIntIndex    = errors.New("need integer index")
	ErrIndexOutOfRange = errors.New("index out of range")
)

type ListLike interface {
	Lener
	Iterator
	IndexOneer
}

// List is a list of Value's.
type List struct {
	inner *[]Value
}

var _ ListLike = List{}

// NewList creates a new List.
func NewList(vs ...Value) List {
	return List{&vs}
}

func (List) Kind() string {
	return "list"
}

func (l List) Repr(indent int) string {
	var b ListReprBuilder
	b.Indent = indent
	for _, v := range *l.inner {
		b.WriteElem(v.Repr(indent + 1))
	}
	return b.String()
}

func (l List) MarshalJSON() ([]byte, error) {
	return json.Marshal(*l.inner)
}

func (l List) Len() int {
	return len(*l.inner)
}

func (l List) Iterate(f func(Value) bool) {
	for _, v := range *l.inner {
		if !f(v) {
			break
		}
	}
}

func (l List) IndexOne(idx Value) Value {
	i := intIndexWithin(idx, len(*l.inner))
	return (*l.inner)[i]
}

func (l List) IndexSet(idx Value, v Value) {
	i := intIndexWithin(idx, len(*l.inner))
	(*l.inner)[i] = v
}

func intIndexWithin(idx Value, n int) int {
	i := intIndex(idx)

	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		throw(ErrIndexOutOfRange)
	}
	return i
}

func intIndex(idx Value) int {
	i, err := strconv.Atoi(ToString(idx))
	if err != nil {
		err := err.(*strconv.NumError)
		if err.Err == strconv.ErrRange {
			throw(ErrIndexOutOfRange)
		} else {
			throw(ErrNeedIntIndex)
		}
	}
	return i
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
