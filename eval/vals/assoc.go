package vals

import (
	"errors"
	"reflect"
)

// Assocer wraps the Assoc method.
type Assocer interface {
	// Assoc returns a slightly modified version of the receiver with key k
	// associated with value v.
	Assoc(k, v interface{}) (interface{}, error)
}

var (
	errAssocUnsupported        = errors.New("assoc is not supported")
	errReplacementMustBeString = errors.New("replacement must be string")
	errAssocWithSlice          = errors.New("assoc with slice not yet supported")
	errStructMapKey            = errors.New("struct-maps must be assoced with existing key")
)

// Assoc takes a container, a key and value, and returns a modified version of
// the container, in which the key associated with the value. It is implemented
// for the builtin type string, List and Map types, StructMap types, and types
// satisfying the Assocer interface. For other types, it returns an error.
func Assoc(a, k, v interface{}) (interface{}, error) {
	switch a := a.(type) {
	case string:
		return assocString(a, k, v)
	case List:
		return assocList(a, k, v)
	case Map:
		return a.Assoc(k, v), nil
	case StructMap:
		return assocStructMap(a, k, v)
	case Assocer:
		return a.Assoc(k, v)
	}
	return nil, errAssocUnsupported
}

func assocString(s string, k, v interface{}) (interface{}, error) {
	i, j, err := convertStringIndex(k, s)
	if err != nil {
		return nil, err
	}
	repl, ok := v.(string)
	if !ok {
		return nil, errReplacementMustBeString
	}
	return s[:i] + repl + s[j:], nil
}

func assocList(l List, k, v interface{}) (interface{}, error) {
	index, err := ConvertListIndex(k, l.Len())
	if err != nil {
		return nil, err
	}
	if index.Slice {
		return nil, errAssocWithSlice
	}
	return l.Assoc(index.Lower, v), nil
}

func assocStructMap(s StructMap, k, v interface{}) (interface{}, error) {
	kstring, ok := k.(string)
	if !ok {
		return nil, errStructMapKey
	}
	t := reflect.TypeOf(s)
	info := getStructMapInfo(t)
	index := -1
	for i, fieldName := range info.fieldNames {
		if fieldName == kstring {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, errStructMapKey
	}
	// TODO: Check whether v can be assigned to the field first, to avoid
	// unncessary copying.
	copy := reflect.New(t).Elem()
	copy.Set(reflect.ValueOf(s))
	err := ScanToGo(v, copy.Field(index).Addr().Interface())
	if err != nil {
		return nil, err
	}
	return copy.Interface(), nil
}
