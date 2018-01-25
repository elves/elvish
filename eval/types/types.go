// Package types contains basic types for the Elvish runtime.
package types

// Definitions for Value interfaces, some simple Value types and some common
// Value helpers.

// Value is an Elvish value.
type Value interface{}

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
	switch v := v.(type) {
	case Stringer:
		return v.String()
	}
	return Repr(v, NoPretty)
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

// Dissocer is anything tha can return a slightly modified version of itself with
// the specified key removed, as a new value.
type Dissocer interface {
	// Dissoc returns a slightly modified version of the receiver with key k
	// dissociated with any value.
	Dissoc(k Value) Value
}
