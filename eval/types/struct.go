package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elves/elvish/parse"
)

var (
	ErrIndexMustBeString = errors.New("index must be string")
)

// Struct is like a Map with fixed keys.
type Struct struct {
	descriptor *StructDescriptor
	fields     []Value
}

var (
	_ Value   = (*Struct)(nil)
	_ MapLike = (*Struct)(nil)
)

// NewStruct creates a new *Struct value.
func NewStruct(descriptor *StructDescriptor, fields []Value) *Struct {
	return &Struct{descriptor, fields}
}

func (*Struct) Kind() string {
	return "map"
}

// Equal returns true if the rhs is MapLike and all pairs are equal.
func (s *Struct) Equal(rhs interface{}) bool {
	return s == rhs || EqMapLike(s, rhs)
}

func (s *Struct) Hash() uint32 {
	return HashMapLike(s)
}

func (s *Struct) Repr(indent int) string {
	var builder MapReprBuilder
	builder.Indent = indent
	for i, name := range s.descriptor.fieldNames {
		builder.WritePair(parse.Quote(name), indent+2, s.fields[i].Repr(indent+2))
	}
	return builder.String()
}

func (s *Struct) Len() int {
	return len(s.descriptor.fieldNames)
}

func (s *Struct) IndexOne(idx Value) (Value, error) {
	i, err := s.index(idx)
	if err != nil {
		return nil, err
	}
	return s.fields[i], nil
}

func (s *Struct) Assoc(k, v Value) Value {
	i, err := s.index(k)
	maybeThrow(err)
	fields := make([]Value, len(s.fields))
	copy(fields, s.fields)
	fields[i] = v
	return &Struct{s.descriptor, fields}
}

func (s *Struct) IterateKey(f func(Value) bool) {
	for _, field := range s.descriptor.fieldNames {
		if !f(String(field)) {
			break
		}
	}
}

func (s *Struct) IteratePair(f func(Value, Value) bool) {
	for i, field := range s.descriptor.fieldNames {
		if !f(String(field), s.fields[i]) {
			break
		}
	}
}

func (s *Struct) HasKey(k Value) bool {
	_, err := s.index(k)
	return err == nil
}

func (s *Struct) index(idx Value) (int, error) {
	index, ok := idx.(String)
	if !ok {
		return 0, ErrIndexMustBeString
	}
	i, ok := s.descriptor.fieldIndex[string(index)]
	if !ok {
		return 0, NoSuchKey(idx)
	}
	return i, nil
}

// MarshalJSON encodes the Struct to a JSON Object.
func (s *Struct) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, fieldName := range s.descriptor.fieldNames {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.Write(s.descriptor.jsonFieldNames[i])
		buf.WriteByte(':')
		fieldJSON, err := json.Marshal(s.fields[i])
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
