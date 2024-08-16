package eval

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/persistent/hash"
)

var (
	// ErrNoOptAccepted is thrown when a Go function that does not accept any
	// options gets passed options.
	ErrNoOptAccepted = errors.New("function does not accept any options")
)

// WrongArgType is thrown when calling a native function with an argument of the
// wrong type.
type WrongArgType struct {
	argNum    int
	typeError error
}

// Error implements the error interface.
func (e WrongArgType) Error() string {
	return fmt.Sprintf("wrong type for arg #%d: %v", e.argNum, e.typeError)
}

// Unwrap returns the wrapped type error.
func (e WrongArgType) Unwrap() error {
	return e.typeError
}

type goFn struct {
	name string
	impl any

	// Type information of impl.

	// If true, pass the frame as a *Frame argument.
	frame bool
	// If true, pass options as a RawOptions argument.
	rawOptions bool
	// If not nil, type of the parameter that gets options via RawOptions.Scan.
	options reflect.Type
	// If not nil, pass the inputs as an Input-typed last argument.
	inputs bool
	// Type of "normal" (non-frame, non-options, non-variadic) arguments.
	normalArgs []reflect.Type
	// If not nil, type of variadic arguments.
	variadicArg reflect.Type
}

// An interface to be implemented by pointers to structs that should hold
// scanned options.
type optionsPtr interface {
	SetDefaultOptions()
}

// Inputs is the type that the last parameter of a Go-native function can take.
// When that is the case, it is a callback to get inputs. See the doc of GoFn
// for details.
type Inputs func(func(any))

var (
	frameType      = reflect.TypeOf((*Frame)(nil))
	rawOptionsType = reflect.TypeOf(RawOptions(nil))
	optionsPtrType = reflect.TypeOf((*optionsPtr)(nil)).Elem()
	inputsType     = reflect.TypeOf(Inputs(nil))
)

// NewGoFn wraps a Go function into an Elvish function using reflection.
//
// Parameters are passed following these rules:
//
// 1. If the first parameter of function has type *Frame, it gets the current
// call frame.
//
// 2. After the potential *Frame argument, the first parameter has type
// RawOptions, it gets a map of option names to their values.
//
// Alternatively, this parameter may be a (non-pointer) struct whose pointer
// type implements a SetDefaultOptions method that takes no arguments and has no
// return value. In this case, a new instance of the struct is constructed, the
// SetDefaultOptions method is called, and any option passed to the Elvish
// function is used to populate the fields of the struct. Field names are mapped
// to option names using strutil.CamelToDashed, unless they have a field tag
// "name", in which case the tag is preferred.
//
// If the function does not declare that it accepts options via either method
// described above, it accepts no options.
//
// 3. If the last parameter is non-variadic and has type Inputs, it represents
// an optional parameter that contains the input to this function. If the
// argument is not supplied, the input channel of the Frame will be used to
// supply the inputs.
//
// 4. Other parameters are converted using vals.ScanToGo.
//
// Return values are written to the stdout channel, after being converted using
// vals.FromGo. Return values whose types are arrays or slices, and not defined
// types, have their individual elements written to the output.
//
// If the last return value has nominal type error and is not nil, it is turned
// into an exception and no return value is written. If the last return value is
// a nil error, it is ignored.
func NewGoFn(name string, impl any) Callable {
	implType := reflect.TypeOf(impl)
	b := &goFn{name: name, impl: impl}

	i := 0
	if i < implType.NumIn() && implType.In(i) == frameType {
		b.frame = true
		i++
	}
	if i < implType.NumIn() && implType.In(i) == rawOptionsType {
		b.rawOptions = true
		i++
	}
	if i < implType.NumIn() && reflect.PointerTo(implType.In(i)).Implements(optionsPtrType) {
		if b.rawOptions {
			panic("Function declares both RawOptions and Options parameters")
		}
		b.options = implType.In(i)
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
func (*goFn) Kind() string {
	return "fn"
}

// Equal compares identity.
func (b *goFn) Equal(rhs any) bool {
	return b == rhs
}

// Hash hashes the address.
func (b *goFn) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(b))
}

// Repr returns an opaque representation "<builtin $name>".
func (b *goFn) Repr(int) string {
	return "<builtin " + b.name + ">"
}

// error(nil) is treated as nil by reflect.TypeOf, so we first get the type of
// *error and use Elem to obtain type of error.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Call calls the implementation using reflection.
func (b *goFn) Call(f *Frame, args []any, opts map[string]any) error {
	if b.variadicArg != nil {
		if len(args) < len(b.normalArgs) {
			return errs.ArityMismatch{What: "arguments",
				ValidLow: len(b.normalArgs), ValidHigh: -1, Actual: len(args)}
		}
	} else if b.inputs {
		if len(args) != len(b.normalArgs) && len(args) != len(b.normalArgs)+1 {
			return errs.ArityMismatch{What: "arguments",
				ValidLow: len(b.normalArgs), ValidHigh: len(b.normalArgs) + 1, Actual: len(args)}
		}
	} else if len(args) != len(b.normalArgs) {
		return errs.ArityMismatch{What: "arguments",
			ValidLow: len(b.normalArgs), ValidHigh: len(b.normalArgs), Actual: len(args)}
	}
	if !b.rawOptions && b.options == nil && len(opts) > 0 {
		return ErrNoOptAccepted
	}

	var in []reflect.Value
	if b.frame {
		in = append(in, reflect.ValueOf(f))
	}
	if b.rawOptions {
		in = append(in, reflect.ValueOf(opts))
	}
	if b.options != nil {
		ptrValue := reflect.New(b.options)
		ptr := ptrValue.Interface()
		ptr.(optionsPtr).SetDefaultOptions()
		err := scanOptions(opts, ptr)
		if err != nil {
			return err
		}
		in = append(in, ptrValue.Elem())
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
			return WrongArgType{i, err}
		}
		in = append(in, ptr.Elem())
	}

	if b.inputs {
		var inputs Inputs
		if len(args) == len(b.normalArgs) {
			inputs = f.IterateInputs
		} else {
			// Wrap an iterable argument in Inputs.
			iterable := args[len(args)-1]
			if !vals.CanIterate(iterable) {
				return fmt.Errorf("%s cannot be iterated", vals.Kind(iterable))
			}
			inputs = func(f func(any)) {
				// CanIterate(iterable) is true
				_ = vals.Iterate(iterable, func(v any) bool {
					f(v)
					return true
				})
			}
		}
		in = append(in, reflect.ValueOf(inputs))
	}

	rets := reflect.ValueOf(b.impl).Call(in)

	if len(rets) > 0 && rets[len(rets)-1].Type() == errorType {
		err := rets[len(rets)-1].Interface()
		if err != nil {
			return err.(error)
		}
		rets = rets[:len(rets)-1]
	}

	out := f.ValueOutput()
	for _, ret := range rets {
		t := ret.Type()
		k := t.Kind()
		if (k == reflect.Slice || k == reflect.Array) && t.Name() == "" {
			for i := 0; i < ret.Len(); i++ {
				err := out.Put(vals.FromGo(ret.Index(i).Interface()))
				if err != nil {
					return err
				}
			}
		} else {
			err := out.Put(vals.FromGo(ret.Interface()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
