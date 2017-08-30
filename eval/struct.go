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

// Equal returns true if the rhs is MapLike and all pairs are equal.
func (s *Struct) Equal(rhs interface{}) bool {
	return s == rhs || eqMapLike(s, rhs)
}

func (s *Struct) Repr(indent int) string {
	var builder MapReprBuilder
	builder.Indent = indent
	for i, name := range s.FieldNames {
		builder.WritePair(parse.Quote(name), indent+2, s.Fields[i].Get().Repr(indent+2))
	}
	return builder.String()
}

func (s *Struct) Len() int {
	return len(s.FieldNames)
}

func (s *Struct) IndexOne(idx Value) Value {
	return s.index(idx).Get()
}

func (s *Struct) IterateKey(f func(Value) bool) {
	for _, field := range s.FieldNames {
		if !f(String(field)) {
			break
		}
	}
}

func (s *Struct) IteratePair(f func(Value, Value) bool) {
	for i, field := range s.FieldNames {
		if !f(String(field), s.Fields[i].Get()) {
			break
		}
	}
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
