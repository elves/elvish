package eval

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/hash"
)

// ReflectBuiltinFn uses reflection to wrap arbitrary Go functions into Elvish
// functions.
//
// If the first parameter of function has type *Frame, it gets the current call
// frame. After that, if the first parameter has type Options, it gets a map of
// options. Parameters of type int or float64 admit all values that can be
// converted using toInt or toFloat. All other parameters admit whatever values
// are assignable to them and no special conversion happens.
//
// If the last return value has type error and is not nil, it is turned into an
// exception. Other return values goes to the channel part of the stdout port.
type ReflectBuiltinFn struct {
	name string
	impl interface{}

	// Type information of impl.

	frame   bool
	options bool
	// Type of "normal" (non-Frame, non-Options, non-variadic) arguments.
	normalArgs []reflect.Type
	// Type of variadic arguments, nil if function is non-variadic
	variadicArg reflect.Type
}

var _ Callable = &ReflectBuiltinFn{}

type Options map[string]interface{}

var (
	frameType   = reflect.TypeOf((*Frame)(nil))
	optionsType = reflect.TypeOf(Options(nil))
)

// NewReflectBuiltinFn creates a new ReflectBuiltinFn instance.
func NewReflectBuiltinFn(name string, impl interface{}) *ReflectBuiltinFn {
	implType := reflect.TypeOf(impl)
	b := &ReflectBuiltinFn{name: name, impl: impl}

	i := 0
	if i < implType.NumIn() && implType.In(i) == frameType {
		b.frame = true
		i++
	}
	if i < implType.NumIn() && implType.In(i) == optionsType {
		b.options = true
		i++
	}
	for ; i < implType.NumIn(); i++ {
		if implType.IsVariadic() && i == implType.NumIn()-1 {
			b.variadicArg = implType.In(i).Elem()
		} else {
			b.normalArgs = append(b.normalArgs, implType.In(i))
		}
	}
	return b
}

// Kind returns "fn".
func (*ReflectBuiltinFn) Kind() string {
	return "fn"
}

// Equal compares identity.
func (b *ReflectBuiltinFn) Equal(rhs interface{}) bool {
	return b == rhs
}

// Hash hashes the address.
func (b *ReflectBuiltinFn) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(b))
}

// Repr returns an opaque representation "<builtin $name>".
func (b *ReflectBuiltinFn) Repr(int) string {
	return "<builtin " + b.name + ">"
}

// error(nil) is treated as nil by reflect.TypeOf, so we first get the type of
// *error and use Elem to obtain type of error.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Call calls the implementation using reflection.
func (b *ReflectBuiltinFn) Call(f *Frame, args []interface{}, opts map[string]interface{}) error {
	if b.variadicArg != nil {
		if len(args) < len(b.normalArgs) {
			return fmt.Errorf("want %d or more arguments, got %d",
				len(b.normalArgs), len(args))
		}
	} else if len(args) != len(b.normalArgs) {
		return fmt.Errorf("want %d arguments, got %d", len(b.normalArgs), len(args))
	}

	var in []reflect.Value
	if b.frame {
		in = append(in, reflect.ValueOf(f))
	}
	if b.options {
		in = append(in, reflect.ValueOf(opts))
	}
	for i, arg := range args {
		var typ reflect.Type
		if i < len(b.normalArgs) {
			typ = b.normalArgs[i]
		} else {
			typ = b.variadicArg
		}
		converted, err := convertArg(arg, typ)
		if err != nil {
			return fmt.Errorf("wrong type of %d'th argument: %v", i+1, err)
		}
		in = append(in, reflect.ValueOf(converted))
	}

	outs := reflect.ValueOf(b.impl).Call(in)

	for i, out := range outs {
		a := out.Interface()
		if i == len(outs)-1 && out.Type() == errorType {
			if a != nil {
				return a.(error)
			}
		} else {
			f.OutputChan() <- a
		}
	}
	return nil
}

var (
	intType   = reflect.TypeOf(int(0))
	floatType = reflect.TypeOf(float64(0))
)

// convertArg converts an argument to the specified type so that it can be
// passed to implementation body. Only conversions to int and float happen
// implicitly; in other cases, the argument must be assignable to the specified
// type.
// TODO: Reimplement scanValueToGo in terms of this function.
func convertArg(arg interface{}, typ reflect.Type) (interface{}, error) {
	switch typ {
	case intType:
		i, err := toInt(arg)
		return i, err
	case floatType:
		f, err := toFloat(arg)
		return f, err
	default:
		if reflect.TypeOf(arg).AssignableTo(typ) {
			return arg, nil
		}
		return nil, fmt.Errorf("need %s, got %s",
			types.Kind(reflect.Zero(typ).Interface()), types.Kind(arg))
	}
}
