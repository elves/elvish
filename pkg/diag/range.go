package diag

// Ranger wraps the Range method.
type Ranger interface {
	// Range returns the range associated with the value.
	Range() Ranging
}

// Ranging represents a range [From, To) within an indexable sequence. Structs
// can embed Ranging to satisfy the [Ranger] interface.
//
// Ideally, this type would be called Range. However, doing that means structs
// embedding this type will have Range as a field instead of a method, thus not
// implementing the [Ranger] interface.
type Ranging struct {
	From int
	To   int
}

// Range returns the Ranging itself.
func (r Ranging) Range() Ranging { return r }

// PointRanging returns a zero-width Ranging at the given point.
func PointRanging(p int) Ranging {
	return Ranging{p, p}
}

// MixedRanging returns a Ranging from the start position of a to the end
// position of b.
func MixedRanging(a, b Ranger) Ranging {
	return Ranging{a.Range().From, b.Range().To}
}
