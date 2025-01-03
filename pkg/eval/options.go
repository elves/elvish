package eval

import (
	"reflect"
	"slices"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

// UnknownOption is thrown by a native function when called with an unknown option.
type UnknownOption struct {
	OptName string
}

// Error implements the error interface.
func (e UnknownOption) Error() string {
	return "unknown option: " + parse.Quote(e.OptName)
}

// RawOptions is the type of an argument a Go-native function can take to
// declare that it wants to parse options itself. See the doc of NewGoFn for
// details.
type RawOptions map[string]any

// Takes a raw option map and a pointer to a field-map struct, populates the
// struct with options.
//
// Similar to vals.ScanToGoOpts(rawOpts, ptr, vals.AllowMissingMapKey), except
// that rawOpts is a Go map rather than an Elvish map.
func scanOptions(rawOpts RawOptions, ptr any) error {
	dstValue := reflect.ValueOf(ptr).Elem()
	keys := vals.GetFieldMapKeys(dstValue.Interface())
	findUnknownOption := func() error {
		for key := range rawOpts {
			if !slices.Contains(keys, key) {
				return UnknownOption{key}
			}
		}
		panic("unreachable")
	}
	if len(rawOpts) > len(keys) {
		return findUnknownOption()
	}
	usedOpts := 0
	for i, key := range keys {
		value, ok := rawOpts[key]
		if !ok {
			continue
		}
		err := vals.ScanToGo(value, dstValue.Field(i).Addr().Interface())
		if err != nil {
			return err
		}
		usedOpts++
	}
	if len(rawOpts) > usedOpts {
		return findUnknownOption()
	}
	return nil
}
