package vals

import (
	"os"

	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
)

// File is an alias for *os.File.
type File = *os.File

// List is an alias for the underlying type used for lists in Elvish.
type List = vector.Vector

// EmptyList is an empty list.
var EmptyList = vector.Empty

// MakeList creates a new List from values.
func MakeList(vs ...interface{}) vector.Vector {
	vec := vector.Empty
	for _, v := range vs {
		vec = vec.Cons(v)
	}
	return vec
}

// MakeStringList creates a new List from strings.
func MakeStringList(vs ...string) vector.Vector {
	vec := vector.Empty
	for _, v := range vs {
		vec = vec.Cons(v)
	}
	return vec
}

// Map is an alias for the underlying type used for maps in Elvish.
type Map = hashmap.Map

// EmptyMap is an empty map.
var EmptyMap = hashmap.New(Equal, Hash)

// MakeMap converts a native Go map to Map.
func MakeMap(raw map[interface{}]interface{}) hashmap.Map {
	m := EmptyMap
	for k, v := range raw {
		m = m.Assoc(k, v)
	}
	return m
}

// MakeMapFromKV creates a map from arguments that are alternately keys and
// values. It panics if the number of arguments is odd.
func MakeMapFromKV(a ...interface{}) hashmap.Map {
	if len(a)%2 == 1 {
		panic("Odd number of arguments to MakeMapFromKV")
	}
	m := EmptyMap
	for i := 0; i < len(a); i += 2 {
		m = m.Assoc(a[i], a[i+1])
	}
	return m
}
