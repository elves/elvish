package eval

import (
	"errors"
	"reflect"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

var ErrNoOptAccepted = errors.New("function does not accept any options")

type RawOptions map[string]interface{}

// Scan takes a pointer to a struct and scan options into it. A field named
// FieldName corresponds to the option named field-name, unless the field has a
// explicit "name" tag. Fields typed ParsedOptions are ignored.
func (rawOpts RawOptions) Scan(ptr interface{}) {
	ptrValue := reflect.ValueOf(ptr)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.Elem().Kind() != reflect.Struct {
		throwf("internal bug: need struct ptr to scan options, got %T", ptr)
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
			throwf("unknown option %s", parse.Quote(k))
		}
		err := vals.ScanToGo(v, struc.Field(fieldIdx).Addr().Interface())
		if err != nil {
			throw(err)
		}
	}
}
