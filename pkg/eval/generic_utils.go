package eval

// Some generic utils that should appear in the standard library soon.

func mapKeys[K comparable, V any](m map[K]V) []K {
	ks := make([]K, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func sliceContains[T comparable](xs []T, y T) bool {
	for _, x := range xs {
		if x == y {
			return true
		}
	}
	return false
}
