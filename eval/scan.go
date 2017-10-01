package eval

// This file implements facilities for "scanning" arguments and options.

import (
	"reflect"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// ScanArgs scans arguments into pointers to supported argument types. If the
// arguments cannot be scanned, an error is thrown.
func ScanArgs(src []Value, dstPtrs ...interface{}) {
	if len(src) != len(dstPtrs) {
		throwf("arity mistmatch: want %d arguments, got %d", len(dstPtrs), len(src))
	}
	for i, value := range src {
		scanValueToGo(value, dstPtrs[i])
	}
}

// ScanArgsVariadic is like ScanArgs, but the last element of args should be a
// pointer to a slice, and the rest of arguments will be scanned into it.
func ScanArgsVariadic(src []Value, dstPtrs ...interface{}) {
	if len(src) < len(dstPtrs)-1 {
		throwf("arity mistmatch: want at least %d arguments, got %d", len(dstPtrs)-1, len(src))
	}
	ScanArgs(src[:len(dstPtrs)-1], dstPtrs[:len(dstPtrs)-1]...)

	// Scan the rest of arguments into a slice.
	rest := src[len(dstPtrs)-1:]
	restDst := reflect.ValueOf(dstPtrs[len(dstPtrs)-1])
	if restDst.Kind() != reflect.Ptr || restDst.Elem().Kind() != reflect.Slice {
		throwf("internal bug: %T to ScanArgsVariadic, need pointer to slice", dstPtrs[len(dstPtrs)-1])
	}
	scanned := reflect.MakeSlice(restDst.Elem().Type(), len(rest), len(rest))
	for i, value := range rest {
		scanValueToGo(value, scanned.Index(i).Addr().Interface())
	}
	reflect.Indirect(restDst).Set(scanned)
}

// ScanArgsOptionalInput is like ScanArgs, but the argument can contain an
// optional iterable value at the end containing inputs to the function. The
// return value is a function that iterates the iterable value if it exists, or
// the input otherwise.
func ScanArgsOptionalInput(ec *EvalCtx, src []Value, dstArgs ...interface{}) func(func(Value)) {
	switch len(src) {
	case len(dstArgs):
		ScanArgs(src, dstArgs...)
		return ec.IterateInputs
	case len(dstArgs) + 1:
		ScanArgs(src[:len(dstArgs)], dstArgs...)
		value := src[len(dstArgs)]
		iterable, ok := value.(Iterable)
		if !ok {
			throwf("need iterable argument, got %s", value.Kind())
		}
		return func(f func(Value)) {
			iterable.Iterate(func(v Value) bool {
				f(v)
				return true
			})
		}
	default:
		throwf("arity mistmatch: want %d or %d arguments, got %d", len(dstArgs), len(dstArgs)+1, len(src))
		return nil
	}
}

// OptToScan is a data structure for an option that is intended to be used in
// ScanOpts.
type OptToScan struct {
	Name    string
	Ptr     interface{}
	Default Value
}

// ScanOpts scans options from a map.
func ScanOpts(m map[string]Value, opts ...OptToScan) {
	scanned := make(map[string]bool)
	for _, opt := range opts {
		a := opt.Ptr
		value, ok := m[opt.Name]
		if !ok {
			value = opt.Default
		}
		scanValueToGo(value, a)
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
func ScanOptsToStruct(m map[string]Value, structPtr interface{}) {
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
		scanValueToGo(v, struc.Field(fieldIdx).Addr().Interface())
	}
}
