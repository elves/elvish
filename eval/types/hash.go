package types

import "github.com/xiaq/persistent/hash"

// Hasher wraps the Hash method.
type Hasher interface {
	// Hash computes the hash code of the receiver.
	Hash() uint32
}

func Hash(v interface{}) uint32 {
	switch v := v.(type) {
	case string:
		return hash.String(v)
	case Hasher:
		return v.Hash()
	}
	return 0
}
