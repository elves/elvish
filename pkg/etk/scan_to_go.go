package etk

import (
	"reflect"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/must"
)

// A slightly nicer wrapper of scanToGo.
func ScanToGo[T any](val any, fm *eval.Frame) (T, error) {
	var dst T
	err := scanToGo(val, &dst, fm)
	if err == nil {
		return dst, nil
	}
	return zero[T](), err
}

// A variant of vals.ScanToGo,
// with additional support for adapting an Elvish function to a Go function,
// or an Elvish map to a Go interface.
func scanToGo(val, ptr any, fm *eval.Frame) error {
	err := vals.ScanToGo(val, ptr)

	dst := reflect.ValueOf(ptr).Elem()
	dstType := reflect.TypeOf(ptr).Elem()
	if fn, ok := val.(eval.Callable); ok && dstType.Kind() == reflect.Func {
		// Adapt an Elvish function to a Go function
		dst.Set(reflect.MakeFunc(dstType, func(args []reflect.Value) []reflect.Value {
			// TODO: Handle errors properly
			// TODO: Add intermediate "internal" entry to the traceback
			outs := must.OK1(fm.CaptureOutput(func(fm *eval.Frame) error {
				return fn.Call(fm, each(args, reflect.Value.Interface), eval.NoOpts)
			}))
			goOuts := make([]reflect.Value, dstType.NumOut())
			if len(outs) != len(goOuts) {
				panic("wrong number of outputs")
			}
			for i, out := range outs {
				goOutPtr := reflect.New(dstType.Out(i))
				must.OK(scanToGo(out, goOutPtr.Interface(), fm))
				goOuts[i] = reflect.Indirect(goOutPtr)
			}
			return goOuts
		}))
		return nil
	} else if structType, ok := structOfFuncForInterface[dstType]; ok {
		// Scan an Elvish map into a Go interface.
		if _, ok := val.(vals.Map); ok {
			// TODO: Accept other map-like types too
			// Create a zero value of the struct type, and fill all of its
			// fields from the map.
			structPtr := reflect.New(structType)
			err := vals.ScanFieldMapFromMap(
				val, structPtr.Interface(),
				vals.GetFieldMapKeys(structPtr.Elem().Interface()),
				vals.AllowExtraMapKey,
				func(src, ptr any, opts vals.ScanOpt) error {
					// TODO: Don't ignore opts
					return scanToGo(src, ptr, fm)
				})
			if err == nil {
				dst.Set(structPtr.Elem())
				return nil
			}
		}
	}
	return err
}

// Maps an interface to its "struct of func" implementation.
var structOfFuncForInterface = map[reflect.Type]reflect.Type{}

// Registers a "struct of func implementation" of an interface.
//
// For example, the following interface:
//
//	type I interface {
//		Foo(a int)
//		Bar() int
//	}
//
// Can be implemented by the following "struct of func":
//
//	type S struct {
//		FooImpl func(a int)
//		BarImpl func() int
//	}
//	func (s S) Foo(a int) { s.FooImpl(a) }
//	func (s S) Bar() int { return s.BarImpl() }
//
// (Each field has the "Impl" by convention.)
//
// And you would call this function like this to register their relationship:
//
//	var _ = RegisterStructOfFuncForInterface[S, I]()
//
// (The function has a useless return value so that it can be called from the top level.)
//
// This registration allows [ScanToGo] to scan an Elvish map into a Go interface,
// like:
//
//	[&foo={|a| ... } & bar={ num 1 }]
func RegisterStructOfFuncForInterface[S, I any]() struct{} {
	stype := reflect.TypeFor[S]()
	itype := reflect.TypeFor[I]()
	if stype.Kind() != reflect.Struct {
		panic("S must be a struct type")
	}
	szero := reflect.Zero(stype).Interface()
	if !vals.IsFieldMap(szero) {
		panic("S must be a field map")
	}
	if itype.Kind() != reflect.Interface {
		panic("I must be an interface type")
	}
	if !stype.AssignableTo(itype) {
		panic("S must implement I")
	}
	structOfFuncForInterface[itype] = stype
	return struct{}{}
}
