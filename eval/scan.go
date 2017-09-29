package eval

// This file implements facilities for "scanning" arguments and options.

import (
	"reflect"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// ScanArgs scans arguments into pointers to supported argument types. If the
// arguments cannot be scanned, an error is thrown.
func ScanArgs(s []Value, args ...interface{}) {
	if len(s) != len(args) {
		throwf("arity mistmatch: want %d arguments, got %d", len(args), len(s))
	}
	for i, value := range s {
		scanArg(value, args[i])
	}
}

// ScanArgsVariadic is like ScanArgs, but the last element of args should be a
// pointer to a slice, and the rest of arguments will be scanned into it.
func ScanArgsVariadic(s []Value, args ...interface{}) {
	if len(s) < len(args)-1 {
		throwf("arity mistmatch: want at least %d arguments, got %d", len(args)-1, len(s))
	}
	ScanArgs(s[:len(args)-1], args[:len(args)-1]...)

	// Scan the rest of arguments into a slice.
	rest := s[len(args)-1:]
	dst := reflect.ValueOf(args[len(args)-1])
	if dst.Kind() != reflect.Ptr || dst.Elem().Kind() != reflect.Slice {
		throwf("internal bug: %T to ScanArgsVariadic, need pointer to slice", args[len(args)-1])
	}
	scanned := reflect.MakeSlice(dst.Elem().Type(), len(rest), len(rest))
	for i, value := range rest {
		scanArg(value, scanned.Index(i).Addr().Interface())
	}
	reflect.Indirect(dst).Set(scanned)
}

// ScanArgsAndOptionalIterate is like ScanArgs, but the argument can contain an
// optional iterable value at the end. The return value is a function that
// iterates the iterable value if it exists, or the input otherwise.
func ScanArgsAndOptionalIterate(ec *EvalCtx, s []Value, args ...interface{}) func(func(Value)) {
	switch len(s) {
	case len(args):
		ScanArgs(s, args...)
		return ec.IterateInputs
	case len(args) + 1:
		ScanArgs(s[:len(args)], args...)
		value := s[len(args)]
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
		throwf("arity mistmatch: want %d or %d arguments, got %d", len(args), len(args)+1, len(s))
		return nil
	}
}

// Opt is a data structure for an option that is intended to be used in ScanOpts.
type Opt struct {
	Name    string
	Ptr     interface{}
	Default Value
}

// ScanOpts scans options from a map.
func ScanOpts(m map[string]Value, opts ...Opt) {
	scanned := make(map[string]bool)
	for _, opt := range opts {
		a := opt.Ptr
		value, ok := m[opt.Name]
		if !ok {
			value = opt.Default
		}
		scanArg(value, a)
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
		scanArg(v, struc.Field(fieldIdx).Addr().Interface())
	}
}

func scanArg(value Value, a interface{}) {
	ptr := reflect.ValueOf(a)
	if ptr.Kind() != reflect.Ptr {
		throwf("internal bug: %T to ScanArgs, need pointer", a)
	}
	v := reflect.Indirect(ptr)
	switch v.Kind() {
	case reflect.Int:
		i, err := toInt(value)
		maybeThrow(err)
		v.Set(reflect.ValueOf(i))
	case reflect.Float64:
		f, err := toFloat(value)
		maybeThrow(err)
		v.Set(reflect.ValueOf(f))
	default:
		if reflect.TypeOf(value).ConvertibleTo(v.Type()) {
			v.Set(reflect.ValueOf(value).Convert(v.Type()))
		} else {
			throwf("need %T argument, got %s", v.Interface(), value.Kind())
		}
	}
}
