package vals

// Dissocer wraps the Dissoc method.
type Dissocer interface {
	// Dissoc returns a slightly modified version of the receiver with key k
	// dissociated with any value.
	Dissoc(k interface{}) interface{}
}

// Dissoc takes a container and a key, and returns a modified version of the
// container, with the given key dissociated with any value. It is implemented
// for the Map type and types satisfying the Dissocer interface. For other
// types, it returns nil.
func Dissoc(a, k interface{}) interface{} {
	switch a := a.(type) {
	case Map:
		return a.Dissoc(k)
	case Dissocer:
		return a.Dissoc(k)
	default:
		return nil
	}
}
