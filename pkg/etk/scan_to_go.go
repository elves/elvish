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
	}
	return err
}
