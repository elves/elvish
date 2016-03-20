package eval

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/parse"
)

var (
	ErrIndexMustBeString = errors.New("index must be string")
)

// Struct is like a Map with fixed keys.
type Struct struct {
	FieldNames []string
	Fields     []Variable
}

var (
	_ Value   = (*Struct)(nil)
	_ MapLike = (*Struct)(nil)
)

func (*Struct) Kind() string {
	return "map"
}

func (s *Struct) Repr(indent int) string {
	var builder MapReprBuilder
	builder.Indent = indent
	for i, name := range s.FieldNames {
		builder.WritePair(parse.Quote(name), s.Fields[i].Get().Repr(indent+1))
	}
	return builder.String()
}

func (s *Struct) Len() int {
	return len(s.FieldNames)
}

func (s *Struct) IndexOne(idx Value) Value {
	return s.index(idx).Get()
}

func (s *Struct) HasKey(k Value) bool {
	index, ok := k.(String)
	if !ok {
		return false
	}
	for _, name := range s.FieldNames {
		if string(index) == name {
			return true
		}
	}
	return false
}

func (s *Struct) IndexSet(idx Value, v Value) {
	s.index(idx).Set(v)
}

func (s *Struct) index(idx Value) Variable {
	index, ok := idx.(String)
	if !ok {
		throw(ErrIndexMustBeString)
	}
	for i, name := range s.FieldNames {
		if string(index) == name {
			return s.Fields[i]
		}
	}
	throw(fmt.Errorf("no such field: %s", index.Repr(NoPretty)))
	panic("unreachable")
}
