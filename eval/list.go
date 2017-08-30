package eval

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/xiaq/persistent/vector"
)

// Error definitions.
var (
	// ErrNeedIntIndex    = errors.New("need integer index")
	ErrBadIndex        = errors.New("bad index")
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrAssocWithSlice  = errors.New("assoc with slice not yet supported")
)

// List is a list of Value's.
type List struct {
	inner vector.Vector
}

// Make sure that List implements ListLike and Assocer at compile time.
var (
	_ ListLike = List{}
	_ Assocer  = List{}
)

// NewList creates a new List.
func NewList(vs ...Value) List {
	vec := vector.Empty
	for _, v := range vs {
		vec = vec.Cons(v)
	}
	return List{vec}
}

func NewListFromVector(vec vector.Vector) List {
	return List{vec}
}

func (List) Kind() string {
	return "list"
}

func (l List) Equal(rhs interface{}) bool {
	return eqListLike(l, rhs)
}

func (l List) Repr(indent int) string {
	var b ListReprBuilder
	b.Indent = indent
	for it := l.inner.Iterator(); it.HasElem(); it.Next() {
		v := it.Elem().(Value)
		b.WriteElem(v.Repr(indent + 1))
	}
	return b.String()
}

func (l List) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	buf.WriteByte('[')
	first := true
	for it := l.inner.Iterator(); it.HasElem(); it.Next() {
		if first {
			first = false
		} else {
			buf.WriteByte(',')
		}
		err := encoder.Encode(it.Elem())
		if err != nil {
			return nil, err
		}
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

func (l List) Len() int {
	return l.inner.Len()
}

func (l List) Iterate(f func(Value) bool) {
	for it := l.inner.Iterator(); it.HasElem(); it.Next() {
		v := it.Elem().(Value)
		if !f(v) {
			break
		}
	}
}

func (l List) IndexOne(idx Value) Value {
	slice, i, j := ParseAndFixListIndex(ToString(idx), l.Len())
	if slice {
		return List{l.inner.SubVector(i, j)}
	}
	return l.inner.Nth(i).(Value)
}

func (l List) Assoc(idx, v Value) Value {
	slice, i, _ := ParseAndFixListIndex(ToString(idx), l.Len())
	if slice {
		throw(ErrAssocWithSlice)
	}
	return List{l.inner.AssocN(i, v)}
}

// ParseAndFixListIndex parses a list index and returns whether the index is a
// slice and "real" (-1 becomes n-1) indicies. It throws errors when the index
// is invalid or out of range.
func ParseAndFixListIndex(s string, n int) (bool, int, int) {
	slice, i, j := parseListIndex(s, n)
	if i < 0 {
		i += n
	}
	if j < 0 {
		j += n
	}
	if i < 0 || i >= n || (slice && (j < 0 || j > n || i > j)) {
		throw(ErrIndexOutOfRange)
	}
	return slice, i, j
}

// ListIndex = Number |
//             Number ':' Number
func parseListIndex(s string, n int) (slice bool, i int, j int) {
	atoi := func(a string) int {
		i, err := strconv.Atoi(a)
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				throw(ErrIndexOutOfRange)
			} else {
				throw(ErrBadIndex)
			}
		}
		return i
	}

	colon := strings.IndexRune(s, ':')
	if colon == -1 {
		// A single number
		return false, atoi(s), 0
	}
	if s[:colon] == "" {
		i = 0
	} else {
		i = atoi(s[:colon])
	}
	if s[colon+1:] == "" {
		j = n
	} else {
		j = atoi(s[colon+1:])
	}
	// Two numbers
	return true, i, j
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
