package types

// Hasher wraps the Hash method.
type Hasher interface {
	// Hash computes the hash code of the receiver.
	Hash() uint32
}

func Hash(v interface{}) uint32 {
	if hasher, ok := v.(Hasher); ok {
		return hasher.Hash()
	}
	return 0
}
