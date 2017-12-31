package eval

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
)

var (
	ErrIndexMustBeString = errors.New("index must be string")
)

// Struct is like a Map with fixed keys.
type Struct struct {
	Descriptor *StructDescriptor
	Fields     []types.Value
}

var (
	_ types.Value = (*Struct)(nil)
	_ MapLike     = (*Struct)(nil)
)

func (*Struct) Kind() string {
	return "map"
}

// Equal returns true if the rhs is MapLike and all pairs are equal.
func (s *Struct) Equal(rhs interface{}) bool {
	return s == rhs || eqMapLike(s, rhs)
}

func (s *Struct) Hash() uint32 {
	return hashMapLike(s)
}

func (s *Struct) Repr(indent int) string {
	var builder MapReprBuilder
	builder.Indent = indent
	for i, name := range s.Descriptor.fieldNames {
		builder.WritePair(parse.Quote(name), indent+2, s.Fields[i].Repr(indent+2))
	}
	return builder.String()
}

func (s *Struct) Len() int {
	return len(s.Descriptor.fieldNames)
}

func (s *Struct) IndexOne(idx types.Value) types.Value {
	return s.Fields[s.index(idx)]
}

func (s *Struct) Assoc(k, v types.Value) types.Value {
	i := s.index(k)
	fields := make([]types.Value, len(s.Fields))
	copy(fields, s.Fields)
	fields[i] = v
	return &Struct{s.Descriptor, fields}
}

func (s *Struct) IterateKey(f func(types.Value) bool) {
	for _, field := range s.Descriptor.fieldNames {
		if !f(String(field)) {
			break
		}
	}
}

func (s *Struct) IteratePair(f func(types.Value, types.Value) bool) {
	for i, field := range s.Descriptor.fieldNames {
		if !f(String(field), s.Fields[i]) {
			break
		}
	}
}

func (s *Struct) HasKey(k types.Value) bool {
	index, ok := k.(String)
	if !ok {
		return false
	}
	_, ok = s.Descriptor.fieldIndex[string(index)]
	return ok
}

func (s *Struct) index(idx types.Value) int {
	index, ok := idx.(String)
	if !ok {
		throw(ErrIndexMustBeString)
	}
	i, ok := s.Descriptor.fieldIndex[string(index)]
	if !ok {
		throw(fmt.Errorf("no such field: %s", index.Repr(types.NoPretty)))
	}
	return i
}

// MarshalJSON encodes the Struct to a JSON Object.
func (s *Struct) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, fieldName := range s.Descriptor.fieldNames {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.Write(s.Descriptor.jsonFieldNames[i])
		buf.WriteByte(':')
		fieldJSON, err := json.Marshal(s.Fields[i])
		if err != nil {
			return nil, fmt.Errorf("cannot encode field %q: %v", fieldName, err)
		}
		buf.Write(fieldJSON)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// StructDescriptor contains information about the fields in a Struct.
type StructDescriptor struct {
	fieldNames     []string
	jsonFieldNames [][]byte
	fieldIndex     map[string]int
}

// NewStructDescriptor creates a new struct descriptor from a list of field
// names.
func NewStructDescriptor(fields ...string) *StructDescriptor {
	fieldNames := append([]string(nil), fields...)
	jsonFieldNames := make([][]byte, len(fields))
	fieldIndex := make(map[string]int)
	for i, name := range fieldNames {
		fieldIndex[name] = i
		jsonFieldName, err := json.Marshal(name)
		// json.Marshal should never fail on string.
		if err != nil {
			panic(err)
		}
		jsonFieldNames[i] = jsonFieldName
	}
	return &StructDescriptor{fieldNames, jsonFieldNames, fieldIndex}
}
