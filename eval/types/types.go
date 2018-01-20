// Package types contains basic types for the Elvish runtime.
package types

import (
	"github.com/elves/elvish/util"
)

// Definitions for Value interfaces, some simple Value types and some common
// Value helpers.

// Value is an Elvish value.
type Value interface {
	Kinder
	Equaler
	Hasher
	Reprer
}

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

// Reprer wraps the Repr method.
type Reprer interface {
	// Repr returns a string that represents a Value. The string either be a
	// literal of that Value that is preferably deep-equal to it (like `[a b c]`
	// for a list), or a string enclosed in "<>" containing the kind and
	// identity of the Value(like `<fn 0xdeadcafe>`).
	//
	// If indent is at least 0, it should be pretty-printed with the current
	// indentation level of indent; the indent of the first line has already
	// been written and shall not be written in Repr. The returned string
	// should never contain a trailing newline.
	Repr(indent int) string
}

// NoPretty can be passed to Repr to suppress pretty-printing.
const NoPretty = util.MinInt

// Equaler wraps the Equal method.
type Equaler interface {
	// Equal compares the receiver to another value. Two equal values must have
	// the same hash code.
	Equal(other interface{}) bool
}

// Hasher wraps the Hash method.
type Hasher interface {
	// Hash computes the hash code of the receiver.
	Hash() uint32
}

// Booler wraps the Bool method.
type Booler interface {
	// Bool computes the truth value of the receiver.
	Bool() bool
}

// Stringer wraps the String method.
type Stringer interface {
	// Stringer converts the receiver to a string.
	String() string
}

// ToString converts a Value to string. When the Value type implements
// String(), it is used. Otherwise Repr(NoPretty) is used.
func ToString(v Value) string {
	if s, ok := v.(Stringer); ok {
		return s.String()
	}
	return v.Repr(NoPretty)
}

// Lener wraps the Len method.
type Lener interface {
	// Len computes the length of the receiver.
	Len() int
}

// Iterator wraps the Iterate method.
type Iterator interface {
	// Iterate calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	Iterate(func(v Value) bool)
}

// IteratorValue is an iterable Value.
type IteratorValue interface {
	Iterator
	Value
}

func CollectFromIterator(it Iterator) []Value {
	var vs []Value
	if lener, ok := it.(Lener); ok {
		vs = make([]Value, 0, lener.Len())
	}
	it.Iterate(func(v Value) bool {
		vs = append(vs, v)
		return true
	})
	return vs
}

// IterateKeyer wraps the IterateKey method.
type IterateKeyer interface {
	// IterateKey calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	IterateKey(func(k Value) bool)
}

// IteratePairer wraps the IteratePair method.
type IteratePairer interface {
	// IteratePair calls the passed function with each key and value within the
	// receiver. The iteration is aborted if the function returns false.
	IteratePair(func(k, v Value) bool)
}

// MultiIndexer wraps the Index method.
type MultiIndexer interface {
	// Index retrieves the values within the receiver at the specified indicies.
	Index(idx []Value) []Value
}

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves one value from the receiver at the specified index.
	Index(idx Value) (Value, error)
}

// GetIndexer adapts a Value to an Indexer if there is an adapter.
func GetIndexer(v Value) (MultiIndexer, bool) {
	if indexer, ok := v.(MultiIndexer); ok {
		return indexer, true
	}
	if indexOneer, ok := v.(Indexer); ok {
		return IndexerIndexer{indexOneer}, true
	}
	return nil, false
}

// MustIndex indexes i with k and returns the value. If the operation
// resulted in an error, it panics. It is useful when the caller knows that the
// key must be present.
func MustIndex(i Indexer, k Value) Value {
	v, err := i.Index(k)
	if err != nil {
		panic(err)
	}
	return v
}

// IndexerIndexer adapts an Indexer to an Indexer by calling all the
// indicies on the Indexr and collect the results.
type IndexerIndexer struct {
	Indexer
}

func (ioi IndexerIndexer) Index(vs []Value) []Value {
	results := make([]Value, len(vs))
	for i, v := range vs {
		var err error
		results[i], err = ioi.Indexer.Index(v)
		maybeThrow(err)
	}
	return results
}

// Assocer wraps the Assoc method.
type Assocer interface {
	// Assoc returns a slightly modified version of the receiver with key k
	// associated with value v.
	Assoc(k, v Value) Value
}

// Dissocer is anything tha can return a slightly modified version of itself with
// the specified key removed, as a new value.
type Dissocer interface {
	// Dissoc returns a slightly modified version of the receiver with key k
	// dissociated with any value.
	Dissoc(k Value) Value
}
