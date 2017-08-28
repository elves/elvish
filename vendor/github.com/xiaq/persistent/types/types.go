// Package types defines types for or related to persistent data structures.
package types

// Equaler is a value that knows whether it is equal to another value or not.
type Equaler interface {
	// Equal returns whether this value is equal to another one.
	Equal(other interface{}) bool
}

// Hasher is a value with a 32-bit hash code.
type Hasher interface {
	// Hash returns the hash for the key. If k1.Equal(k2), k1.Hash() ==
	// k2.Hash() must be satisfied.
	Hash() uint32
}

// EqualHasher packs Hasher and Equaler.
type EqualHasher interface {
	Equaler
	Hasher
}
