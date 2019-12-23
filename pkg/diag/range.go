package diag

// Ranger wraps the Range method.
type Ranger interface {
	// Range returns the range associated with the value.
	Range() Ranging
}

// Ranging represents a range [From, To) within an indexable sequence. Structs
// can embed Ranging to satisfy the Ranger interface.
type Ranging struct {
	From int
	To   int
}

// Range returns the Ranging itself.
func (r Ranging) Range() Ranging { return r }
