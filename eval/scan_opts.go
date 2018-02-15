package eval

// This file implements facilities for "scanning" arguments and options.

import (
	"errors"
	"reflect"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// OptToScan is a data structure for an option that is intended to be used in
// ScanOpts.
type OptToScan struct {
	Name    string
	Ptr     interface{}
	Default interface{}
}

// ScanOpts scans options from a map.
func ScanOpts(m map[string]interface{}, opts ...OptToScan) {
	scanned := make(map[string]bool)
	for _, opt := range opts {
		a := opt.Ptr
		value, ok := m[opt.Name]
		if !ok {
			value = opt.Default
		}
		mustScanToGo(value, a)
		scanned[opt.Name] = true
	}
	for key := range m {
		if !scanned[key] {
			throwf("unknown option %s", parse.Quote(key))
		}
	}
}

// ScanOptsToStruct scan options from a map like ScanOpts except the destination
// is a struct whose fields correspond to the options to be parsed. A field
// named FieldName corresponds to the option named field-name, unless the field
// has a explicit "name" tag.
func ScanOptsToStruct(m map[string]interface{}, structPtr interface{}) {
	ptrValue := reflect.ValueOf(structPtr)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.Elem().Kind() != reflect.Struct {
		throwf("internal bug: need struct ptr for ScanOptsToStruct, got %T", structPtr)
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

	for k, v := range m {
		fieldIdx, ok := fieldIdxForOpt[k]
		if !ok {
			throwf("unknown option %s", parse.Quote(k))
		}
		mustScanToGo(v, struc.Field(fieldIdx).Addr().Interface())
	}
}

var ErrNoOptAccepted = errors.New("no option accepted")

func TakeNoOpt(opts map[string]interface{}) {
	if len(opts) > 0 {
		throw(ErrNoOptAccepted)
	}
}
