package mtd

import (
	"fmt"
	"reflect"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/strutil"
)

type Method[FT any] struct {
	Call FT
}

var (
	anyType   = reflect.TypeFor[any]()
	errorType = reflect.TypeFor[error]()
	frameType = reflect.TypeFor[*eval.Frame]()
)

func New[FT any](goName string) Method[FT] {
	ft := reflect.TypeFor[FT]()
	if ft.Kind() != reflect.Func {
		panic("FT must be a function type")
	}
	if ft.NumIn() < 2 || ft.In(0) != frameType || ft.In(1) != anyType {
		panic("the first two argument of FT must be *eval.Frame and any")
	}
	if ft.NumOut() == 0 || ft.Out(ft.NumOut()-1) != errorType {
		panic("the last result of FT must be error")
	}

	elvishName := strutil.CamelToDashed(goName)

	callImpl := func(args []reflect.Value) []reflect.Value {
		fm := args[0].Interface().(*eval.Frame)
		// Important: receiver has type any,
		// so we need to "unwrap" it to get the actual receiver value;
		// otherwise the method lookup will fail.
		receiver := args[1].Elem()
		// TODO:
		// If we have an explicitly registered implementation,
		// use it.

		// If the receiver is an Elvish value with a corresponding key,
		// call the implementation.
		if v, err := vals.Index(receiver.Interface(), elvishName); err == nil {
			if fn, ok := v.(eval.Callable); ok {
				fnArgs := make([]any, len(args)-2)
				for i := range fnArgs {
					fnArgs[i] = vals.FromGo(args[i+2].Interface())
				}
				elvishOuts, err := fm.CaptureOutput(func(fm *eval.Frame) error {
					return fn.Call(fm, fnArgs, eval.NoOpts)
				})
				if err != nil {
					return makeOutsWithError(ft, err)
				}
				if len(elvishOuts) != ft.NumOut()-1 {
					return makeOutsWithError(ft, errs.ArityMismatch{
						What:      "output of method implementation",
						ValidLow:  ft.NumOut() - 1,
						ValidHigh: ft.NumOut() - 1,
						Actual:    len(elvishOuts),
					})
				}
				goOuts := make([]reflect.Value, len(elvishOuts)+1)
				for i, out := range elvishOuts {
					ptrGoOut := reflect.New(ft.Out(i))
					vals.ScanToGo(out, ptrGoOut.Interface())
					goOuts[i] = ptrGoOut.Elem()
				}
				goOuts[len(goOuts)-1] = reflect.Zero(errorType)
				return goOuts
			}
		}

		// If receiver implements a matching native Go method, call it.
		// Also accept an implementation that lacks the final error result.
		//
		// NOTE: It's important that this comes after the Elvish key check:
		// this allows Elvish maps to implement methods with the same name as
		// a native Map method.
		goMethod := receiver.MethodByName(goName)
		if goMethod.IsValid() && methodMatches(goMethod.Type(), ft) {
			outs := goMethod.Call(args[2:])
			if len(outs) == ft.NumOut()-1 {
				// Add a nil error.
				outs = append(outs, reflect.Zero(errorType))
			}
			return outs
		}

		// No matching implementation found. Return zero values and an error.
		return makeOutsWithError(ft, fmt.Errorf("no implementation of %s", goName))
	}
	call := reflect.MakeFunc(reflect.TypeFor[FT](), callImpl).Interface().(FT)
	return Method[FT]{ /*impls: impls,*/ Call: call}
}

func makeOutsWithError(ft reflect.Type, err error) []reflect.Value {
	outs := make([]reflect.Value, ft.NumOut())
	for i := range ft.NumOut() - 1 {
		outs[i] = reflect.Zero(ft.Out(i))
	}
	outs[ft.NumOut()-1] = reflect.ValueOf(err)
	return outs
}

// func (m *Method[FT]) Add(t reflect.Type, impl FT) { m.impls[t] = impl }

func methodMatches(mt, ft reflect.Type) bool {
	// mt is the bound method's type,
	// so it doesn't have the initial receiver.
	if mt.NumIn() != ft.NumIn()-2 {
		return false
	}
	// Allow the Go method to have one fewer return value (error always nil).
	if mt.NumOut() < ft.NumOut()-1 || mt.NumOut() > ft.NumOut() {
		return false
	}
	for i := range mt.NumIn() {
		if mt.In(i) != ft.In(i+2) {
			return false
		}
	}
	for i := range mt.NumOut() {
		if mt.Out(i) != ft.Out(i) {
			return false
		}
	}
	return true
}
