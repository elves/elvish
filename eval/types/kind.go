package types

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

func Kind(v interface{}) string {
	if kinder, ok := v.(Kinder); ok {
		return kinder.Kind()
	}
	return "unknown"
}
