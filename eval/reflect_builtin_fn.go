package eval

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/hash"
)

var reflectBuiltinFns = map[string]interface{}{}

func addToReflectBuiltinFns(moreFns map[string]interface{}) {
	for name, impl := range moreFns {
		reflectBuiltinFns[name] = impl
	}
}

// ReflectBuiltinFn uses reflection to wrap arbitrary Go functions into Elvish
// functions.
//
// Parameters are passed following these rules:
//
// 1. If the first parameter of function has type *Frame, it gets the current call
//    frame.
//
// 2. If (possibly after a *Frame parameter) the first parameter has type
//    Options, it gets a map of options. If the function has not declared an
//    Options parameter but is passed options, an error is thrown.
//
// 3. If the last parameter is non-variadic and has type Inputs, it represents
//    an optional parameter that contains the input to this function. If the
//    argument is not supplied, the input channel of the Frame will be used to
//    supply the inputs.
//
// 4. Other parameters are converted using elvToGo.
//
// Return values go to the channel part of the stdout port, after being
// converted using goToElv. If the last return value has type error and is not
// nil, it is turned into an exception and no ouputting happens. If the last
// return value is a nil error, it is ignored.
type ReflectBuiltinFn struct {
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

var _ Callable = &ReflectBuiltinFn{}

type (
	Options map[string]interface{}
	Inputs  func(func(interface{}))
)

func (opt Options) Scan(opts ...OptToScan) {
	ScanOpts(map[string]interface{}(opt), opts...)
}

func (opt Options) ScanToStruct(ptr interface{}) {
	ScanOptsToStruct(map[string]interface{}(opt), ptr)
}

var (
	frameType   = reflect.TypeOf((*Frame)(nil))
	optionsType = reflect.TypeOf(Options(nil))
	inputsType  = reflect.TypeOf(Inputs(nil))
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

var errNoOptions = errors.New("function does not accept any options")

// Call calls the implementation using reflection.
func (b *ReflectBuiltinFn) Call(f *Frame, args []interface{}, opts map[string]interface{}) error {
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
		converted, err := elvToGo(arg, typ)
		if err != nil {
			return fmt.Errorf("wrong type of %d'th argument: %v", i+1, err)
		}
		in = append(in, reflect.ValueOf(converted))
	}

	if b.inputs {
		var inputs Inputs
		if len(args) == len(b.normalArgs) {
			inputs = Inputs(f.IterateInputs)
		} else {
			// Wrap an iterable argument in Inputs.
			iterable := args[len(args)-1]
			inputs = Inputs(func(f func(interface{})) {
				err := types.Iterate(iterable, func(v interface{}) bool {
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
		f.OutputChan() <- goToElv(out.Interface())
	}
	return nil
}
