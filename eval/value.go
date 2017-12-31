package eval

import (
	"fmt"

	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
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

// Lener wraps the Len method.
type Lener interface {
	// Len computes the length of the receiver.
	Len() int
}

// Iterable wraps the Iterate method.
type Iterable interface {
	// Iterate calls the passed function with each value within the receiver.
	// The iteration is aborted if the function returns false.
	Iterate(func(v Value) bool)
}

// IterableValue is an iterable Value.
type IterableValue interface {
	Iterable
	Value
}

func collectFromIterable(it Iterable) []Value {
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

// Callable wraps the Call method.
type Callable interface {
	// Call calls the receiver in a Frame with arguments and options.
	Call(ec *Frame, args []Value, opts map[string]Value)
}

var (
	// NoArgs is an empty argument list. It can be used as an argument to Call.
	NoArgs = []Value{}
	// NoOpts is an empty option map. It can be used as an argument to Call.
	NoOpts = map[string]Value{}
)

// Fn is a callable value.
type Fn interface {
	Value
	Callable
}

// Indexer wraps the Index method.
type Indexer interface {
	// Index retrieves the values within the receiver at the specified indicies.
	Index(idx []Value) []Value
}

// IndexOneer wraps the IndexOne method.
type IndexOneer interface {
	// Index retrieves one value from the receiver at the specified index.
	IndexOne(idx Value) Value
}

func mustIndexer(v Value, ec *Frame) Indexer {
	indexer, ok := getIndexer(v, ec)
	if !ok {
		throw(fmt.Errorf("a %s is not indexable", v.Kind()))
	}
	return indexer
}

// getIndexer adapts a Value to an Indexer if there is an adapter. It adapts a
// Fn if ec is not nil.
func getIndexer(v Value, ec *Frame) (Indexer, bool) {
	if indexer, ok := v.(Indexer); ok {
		return indexer, true
	}
	if indexOneer, ok := v.(IndexOneer); ok {
		return IndexOneerIndexer{indexOneer}, true
	}
	return nil, false
}

// IndexOneerIndexer adapts an IndexOneer to an Indexer by calling all the
// indicies on the IndexOner and collect the results.
type IndexOneerIndexer struct {
	IndexOneer
}

func (ioi IndexOneerIndexer) Index(vs []Value) []Value {
	results := make([]Value, len(vs))
	for i, v := range vs {
		results[i] = ioi.IndexOneer.IndexOne(v)
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

// IndexOneAssocer combines IndexOneer and Assocer.
type IndexOneAssocer interface {
	IndexOneer
	Assocer
}

// FromJSONInterface converts a interface{} that results from json.Unmarshal to
// a Value.
func FromJSONInterface(v interface{}) Value {
	if v == nil {
		// TODO Use a more appropriate type
		return String("")
	}
	switch v.(type) {
	case bool:
		return Bool(v.(bool))
	case float64, string:
		// TODO Use a numeric type for float64
		return String(fmt.Sprint(v))
	case []interface{}:
		a := v.([]interface{})
		vs := make([]Value, len(a))
		for i, v := range a {
			vs[i] = FromJSONInterface(v)
		}
		return NewList(vs...)
	case map[string]interface{}:
		m := v.(map[string]interface{})
		mv := hashmap.Empty
		for k, v := range m {
			mv = mv.Assoc(String(k), FromJSONInterface(v))
		}
		return NewMap(mv)
	default:
		throw(fmt.Errorf("unexpected json type: %T", v))
		return nil // not reached
	}
}
