package hashmap

// Equaler is a value that knows whether it is equal to another value or not.
type Equaler interface {
	// Equal returns whether this value is equal to another one.
	Equal(other interface{}) bool
}

// Key is the interface hashmap keys need to satisfy.
type Key interface {
	Equaler
	// Hash returns the hash for the key. If k1.Equal(k2), k1.Hash() ==
	// k2.Hash() must be satisfied.
	Hash() uint32
}

// UInt32 is a uint32 that can be used as a Key.
type UInt32 uint32

var _ Key = UInt32(0)

func (i UInt32) Hash() uint32 {
	return uint32(i)
}

func (i UInt32) Equal(other interface{}) bool {
	j, ok := other.(UInt32)
	return ok && i == j
}

// UInt64 is a uint64 that can be used as a Key.
type UInt64 uint64

var _ Key = UInt64(0)

// Hash returns the hash code of an UInt64.
func (i UInt64) Hash() uint32 {
	return uint32(i>>32)*33 + uint32(i&0xffffffff)
}

// Equal returns whether a Int64 is equal to another value.
func (i UInt64) Equal(other interface{}) bool {
	j, ok := other.(UInt64)
	return ok && i == j
}

// String is a String that can be used as a Key.
type String string

var _ Key = String("")

// Hash returns the hash code of a String. It uses the djb hashing algorithm.
func (s String) Hash() uint32 {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return h
}

// Equal returns whether a String is equal to another value.
func (s String) Equal(other interface{}) bool {
	t, ok := other.(String)
	return ok && s == t
}
