package vals

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

var (
	ErrIndexMustBeString = errors.New("index must be string")
)

// Struct is like a Map with fixed keys.
type Struct struct {
	descriptor *StructDescriptor
	fields     []interface{}
}

var _ Indexer = (*Struct)(nil)

// NewStruct creates a new *Struct value.
func NewStruct(descriptor *StructDescriptor, fields []interface{}) *Struct {
	return &Struct{descriptor, fields}
}

func (*Struct) Kind() string {
	return "map"
}

// Equal returns true if the rhs is MapLike and all pairs are equal.
func (s *Struct) Equal(rhs interface{}) bool {
	if s == rhs {
		return true
	}
	s2, ok := rhs.(*Struct)
	if !ok {
		return false
	}
	if s.descriptor != s2.descriptor {
		return false
	}
	for i, field := range s.fields {
		if !Equal(field, s2.fields[i]) {
			return false
		}
	}
	return true
}

func (s *Struct) Hash() uint32 {
	h := hash.DJBInit
	for _, field := range s.fields {
		h = hash.DJBCombine(h, Hash(field))
	}
	return h
}

func (s *Struct) Repr(indent int) string {
	var builder MapReprBuilder
	builder.Indent = indent
	for i, name := range s.descriptor.fieldNames {
		builder.WritePair(parse.Quote(name), indent+2, Repr(s.fields[i], indent+2))
	}
	return builder.String()
}

func (s *Struct) Len() int {
	return len(s.descriptor.fieldNames)
}

func (s *Struct) Index(k interface{}) (interface{}, bool) {
	i, ok := s.index(k)
	if !ok {
		return nil, false
	}
	return s.fields[i], true
}

func (s *Struct) Assoc(k, v interface{}) (interface{}, error) {
	i, ok := s.index(k)
	if !ok {
		return nil, NoSuchKey(k)
	}
	fields := make([]interface{}, len(s.fields))
	copy(fields, s.fields)
	fields[i] = v
	return &Struct{s.descriptor, fields}, nil
}

func (s *Struct) IterateKey(f func(interface{}) bool) {
	for _, field := range s.descriptor.fieldNames {
		if !f(string(field)) {
			break
		}
	}
}

func (s *Struct) IteratePair(f func(interface{}, interface{}) bool) {
	for i, field := range s.descriptor.fieldNames {
		if !f(string(field), s.fields[i]) {
			break
		}
	}
}

func (s *Struct) HasKey(k interface{}) bool {
	_, ok := s.index(k)
	return ok
}

func (s *Struct) index(k interface{}) (int, bool) {
	kstring, ok := k.(string)
	if !ok {
		return 0, false
	}
	index, ok := s.descriptor.fieldIndex[kstring]
	return index, ok
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
