package hashmap

import (
	"github.com/xiaq/persistent/hash"
	"github.com/xiaq/persistent/types"
)

// UInt32 is a uint32 that can be used as a types.Key.
type UInt32 uint32

var _ types.EqualHasher = UInt32(0)

func (i UInt32) Hash() uint32 {
	return uint32(i)
}

func (i UInt32) Equal(other interface{}) bool {
	j, ok := other.(UInt32)
	return ok && i == j
}

// UInt64 is a uint64 that can be used as a types.Key.
type UInt64 uint64

var _ types.EqualHasher = UInt64(0)

// Hash returns the hash code of an UInt64.
func (i UInt64) Hash() uint32 {
	return hash.UInt64(uint64(i))
}

// Equal returns whether a Int64 is equal to another value.
func (i UInt64) Equal(other interface{}) bool {
	j, ok := other.(UInt64)
	return ok && i == j
}

// String is a String that can be used as a types.Key.
type String string

var _ types.EqualHasher = String("")

// Hash returns the hash code of a String. It uses the djb hashing algorithm.
func (s String) Hash() uint32 {
	return hash.String(string(s))
}

// Equal returns whether a String is equal to another value.
func (s String) Equal(other interface{}) bool {
	t, ok := other.(String)
	return ok && s == t
}
