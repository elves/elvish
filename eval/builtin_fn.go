package eval

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/elves/elvish/eval/vals"
	"github.com/xiaq/persistent/hash"
)

var ErrArgs = errors.New("args error")

// BuiltinFn uses reflection to wrap arbitrary Go functions into Elvish
// functions.
//
// Parameters are passed following these rules:
//
// 1. If the first parameter of function has type *Frame, it gets the current
// call frame.
//
// 2. If (possibly after a *Frame parameter) the first parameter has type
// RawOptions, it gets a map of options. If the function has not declared an
// RawOptions parameter but is passed options, an error is thrown.
//
// 3. If the last parameter is non-variadic and has type Inputs, it represents
// an optional parameter that contains the input to this function. If the
// argument is not supplied, the input channel of the Frame will be used to
// supply the inputs.
//
// 4. Other parameters are converted using elvToGo.
//
// Return values go to the channel part of the stdout port, after being
// converted using goToElv. If the last return value has type error and is not
// nil, it is turned into an exception and no ouputting happens. If the last
// return value is a nil error, it is ignored.
type BuiltinFn struct {
	name string
	impl interface{}

	// Type information of impl.

	frame   bool
	options bool
	inputs  bool
	// Type of "normal" (non-Frame, non-Options, non-variadic) arguments.
	normalArgs []reflect.Type
	// Type of variadic arguments, nil if function is non-variadic
	variadicArg reflect.Type
}

var _ Callable = &BuiltinFn{}

type (
	Inputs func(func(interface{}))
)

var (
	frameType      = reflect.TypeOf((*Frame)(nil))
	rawOptionsType = reflect.TypeOf(RawOptions(nil))
	inputsType     = reflect.TypeOf(Inputs(nil))
)

// NewBuiltinFn creates a new ReflectBuiltinFn instance.
func NewBuiltinFn(name string, impl interface{}) *BuiltinFn {
	implType := reflect.TypeOf(impl)
	b := &BuiltinFn{name: name, impl: impl}

	i := 0
	if i < implType.NumIn() && implType.In(i) == frameType {
		b.frame = true
		i++
	}
	if i < implType.NumIn() && implType.In(i) == rawOptionsType {
		b.options = true
		i++
	}
	for ; i < implType.NumIn(); i++ {
		paramType := implType.In(i)
		if i == implType.NumIn()-1 {
			if implType.IsVariadic() {
				b.variadicArg = paramType.Elem()
				break
			} else if paramType == inputsType {
				b.inputs = true
				break
			}
		}
		b.normalArgs = append(b.normalArgs, paramType)
	}
	return b
}

// Kind returns "fn".
func (*BuiltinFn) Kind() string {
	return "fn"
}

// Equal compares identity.
func (b *BuiltinFn) Equal(rhs interface{}) bool {
	return b == rhs
}

// Hash hashes the address.
func (b *BuiltinFn) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(b))
}

// Repr returns an opaque representation "<builtin $name>".
func (b *BuiltinFn) Repr(int) string {
	return "<builtin " + b.name + ">"
}

// error(nil) is treated as nil by reflect.TypeOf, so we first get the type of
// *error and use Elem to obtain type of error.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

var errNoOptions = errors.New("function does not accept any options")

// Call calls the implementation using reflection.
func (b *BuiltinFn) Call(f *Frame, args []interface{}, opts map[string]interface{}) error {
	if b.variadicArg != nil {
		if len(args) < len(b.normalArgs) {
			return fmt.Errorf("want %d or more arguments, got %d",
				len(b.normalArgs), len(args))
		}
	} else if b.inputs {
		if len(args) != len(b.normalArgs) && len(args) != len(b.normalArgs)+1 {
			return fmt.Errorf("want %d or %d arguments, got %d",
				len(b.normalArgs), len(b.normalArgs)+1, len(args))
		}
	} else if len(args) != len(b.normalArgs) {
		return fmt.Errorf("want %d arguments, got %d", len(b.normalArgs), len(args))
	}
	if !b.options && len(opts) > 0 {
		return errNoOptions
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
		} else if b.variadicArg != nil {
			typ = b.variadicArg
		} else if b.inputs {
			break // Handled after the loop
		} else {
			panic("impossible")
		}
		ptr := reflect.New(typ)
		err := vals.ScanToGo(arg, ptr.Interface())
		if err != nil {
			return fmt.Errorf("wrong type of %d'th argument: %v", i+1, err)
		}
		in = append(in, ptr.Elem())
	}

	if b.inputs {
		var inputs Inputs
		if len(args) == len(b.normalArgs) {
			inputs = Inputs(f.IterateInputs)
		} else {
			// Wrap an iterable argument in Inputs.
			iterable := args[len(args)-1]
			inputs = Inputs(func(f func(interface{})) {
				err := vals.Iterate(iterable, func(v interface{}) bool {
					f(v)
					return true
				})
				maybeThrow(err)
			})
		}
		in = append(in, reflect.ValueOf(inputs))
	}

	outs := reflect.ValueOf(b.impl).Call(in)

	if len(outs) > 0 && outs[len(outs)-1].Type() == errorType {
		err := outs[len(outs)-1].Interface()
		if err != nil {
			return err.(error)
		}
		outs = outs[:len(outs)-1]
	}

	for _, out := range outs {
		f.OutputChan() <- vals.FromGo(out.Interface())
	}
	return nil
}
