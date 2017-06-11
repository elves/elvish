package hashmap

// Key is the interface hashmap keys need to satisfy.
type Key interface {
	// Hash returns the hash for the key. If k1.Equal(k2), k1.Hash() ==
	// k2.Hash() must be satisfied.
	Hash() uint32
	// Equal returns whether this key is equal to another one.
	Equal(other interface{}) bool
}

// UInt64 is a uint64 that can be used as a Key.
type UInt64 uint64

var _ Key = UInt64(0)

// Hash returns the hash code of an Int64.
//
// TODO(xiaq): Use an actual hashing algorithm.
func (i UInt64) Hash() uint32 {
	return uint32((i >> 32) ^ (i & 0xffffffff))
}

// Equal returns whether a Int64 is equal to another value.
func (i UInt64) Equal(other interface{}) bool {
	j, ok := other.(UInt64)
	return ok && i == j
}

// String is a String that can be used as a Key.
type String string

var _ Key = String("")

// Hash returns the hash code of a String.
//
// TODO(xiaq): Use an actual hashing algorithm.
func (s String) Hash() uint32 {
	return uint32(len(s))
}

// Equal returns whether a String is equal to another value.
func (s String) Equal(other interface{}) bool {
	t, ok := other.(String)
	return ok && s == t
}
