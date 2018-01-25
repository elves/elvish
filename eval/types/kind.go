package types

import "fmt"

// Kinder wraps the Kind method.
type Kinder interface {
	Kind() string
}

func Kind(v interface{}) string {
	switch v := v.(type) {
	case string:
		return "string"
	case Kinder:
		return v.Kind()
	default:
		return fmt.Sprintf("!!%T", v)
	}
}
