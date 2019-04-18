package eval

import (
	"fmt"
	"reflect"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// RawOptions is the type of an argument a Go-native function can take to
// declare that it wants to parse options itself. See the doc of GoFn for
// details.
type RawOptions map[string]interface{}

// Takes a raw option map and a pointer to a struct, and populate the struct
// with options. A field named FieldName corresponds to the option named
// field-name, unless the field has a explicit "name" tag. Fields typed
// ParsedOptions are ignored.
func scanOptions(rawOpts RawOptions, ptr interface{}) error {
	ptrValue := reflect.ValueOf(ptr)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf(
			"internal bug: need struct ptr to scan options, got %T", ptr)
	}
	struc := ptrValue.Elem()

	// fieldIdxForOpt maps option name to the index of field in struc.
	fieldIdxForOpt := make(map[string]int)
	for i := 0; i < struc.Type().NumField(); i++ {
		// ignore unexported fields
		if !struc.Field(i).CanSet() {
			continue
		}

		f := struc.Type().Field(i)
		optName := f.Tag.Get("name")
		if optName == "" {
			optName = util.CamelToDashed(f.Name)
		}
		fieldIdxForOpt[optName] = i
	}

	for k, v := range rawOpts {
		fieldIdx, ok := fieldIdxForOpt[k]
		if !ok {
			return fmt.Errorf("unknown option %s", parse.Quote(k))
		}
		err := vals.ScanToGo(v, struc.Field(fieldIdx).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}
