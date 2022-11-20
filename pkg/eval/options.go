package eval

import (
	"reflect"

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

// Takes a raw option map and a pointer to a struct, and populate the struct
// with options. A field named FieldName corresponds to the option named
// field-name. Options that don't have corresponding fields in the struct causes
// an error.
//
// Similar to vals.ScanMapToGo, but requires rawOpts to contain a subset of keys
// supported by the struct.
func scanOptions(rawOpts RawOptions, ptr any) error {
	_, keyIdx := vals.StructFieldsInfo(reflect.TypeOf(ptr).Elem())
	structValue := reflect.ValueOf(ptr).Elem()
	for k, v := range rawOpts {
		fieldIdx, ok := keyIdx[k]
		if !ok {
			return UnknownOption{k}
		}

		// An option with no value (e.g. `&a-opt`) has `$true` as its default value. However, if
		// the option struct member is a string we want an empty string as the default value.
		switch b := v.(type) {
		case bool:
			if b && structValue.Field(fieldIdx).Type().Name() == "string" {
				v = ""
			}
		}

		err := vals.ScanToGo(v, structValue.Field(fieldIdx).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}
