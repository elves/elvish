package eval

import (
	"bytes"
	"errors"
)

// Map is a map from string to Value.
type Map struct {
	inner *map[Value]Value
}

type MapLike interface {
	Lener
	IndexOneer
}

// NewMap creates a new Map.
func NewMap(inner map[Value]Value) Map {
	return Map{&inner}
}

func (Map) Kind() string {
	return "map"
}

func (m Map) Repr() string {
	var builder MapReprBuilder
	for k, v := range *m.inner {
		builder.WritePair(k.Repr(), v.Repr())
	}
	return builder.String()
}

func (m Map) Len() int {
	return len(*m.inner)
}

func (m Map) IndexOne(idx Value) Value {
	v, ok := (*m.inner)[idx]
	if !ok {
		throw(errors.New("no such key: " + idx.Repr()))
	}
	return v
}

func (m Map) IndexSet(idx Value, v Value) {
	(*m.inner)[idx] = v
}

// MapReprBuilder helps building the Repr of a Map. It is also useful for
// implementing other Map-like values. The zero value of a MapReprBuilder is
// ready to use.
type MapReprBuilder struct {
	buf bytes.Buffer
}

func (b *MapReprBuilder) WritePair(k, v string) {
	if b.buf.Len() == 0 {
		b.buf.WriteByte('[')
	} else {
		b.buf.WriteByte(' ')
	}
	b.buf.WriteString("&" + k + " " + v)
}

func (b *MapReprBuilder) String() string {
	if b.buf.Len() == 0 {
		return "[&]"
	}
	b.buf.WriteByte(']')
	return b.buf.String()
}
