package vals

import (
	"os"

	"src.elv.sh/pkg/persistent/hashmap"
	"src.elv.sh/pkg/persistent/vector"
)

// File is an alias for *os.File.
type File = *os.File

// List is an alias for the underlying type used for lists in Elvish.
type List = vector.Vector

// EmptyList is an empty list.
var EmptyList = vector.Empty

// MakeList creates a new List from values.
func MakeList(vs ...any) vector.Vector {
	return MakeListSlice(vs)
}

// MakeListSlice creates a new List from a slice.
func MakeListSlice[T any](vs []T) vector.Vector {
	vec := vector.Empty
	for _, v := range vs {
		vec = vec.Conj(v)
	}
	return vec
}

// Map is an alias for the underlying type used for maps in Elvish.
type Map = hashmap.Map

// EmptyMap is an empty map.
var EmptyMap = hashmap.New(Equal, Hash)

// MakeMap creates a map from arguments that are alternately keys and values. It
// panics if the number of arguments is odd.
func MakeMap(a ...any) hashmap.Map {
	if len(a)%2 == 1 {
		panic("odd number of arguments to MakeMap")
	}
	m := EmptyMap
	for i := 0; i < len(a); i += 2 {
		m = m.Assoc(a[i], a[i+1])
	}
	return m
}
