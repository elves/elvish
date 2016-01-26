package eval

// Builtin functions.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"reflect"
	"strconv"
)

var builtinFns []*builtinFn

func init() {
	// Needed to work around init loop.
	builtinFns = []*builtinFn{
		&builtinFn{":", nop},
		&builtinFn{"true", nop},

		&builtinFn{"print", wrapFn(print)},
		&builtinFn{"println", wrapFn(println)},

		&builtinFn{"into-lines", wrapFn(intoLines)},
		&builtinFn{"from-lines", wrapFn(fromLines)},

		&builtinFn{"rat", wrapFn(ratFn)},

		&builtinFn{"put", put},
		&builtinFn{"unpack", wrapFn(unpack)},

		&builtinFn{"from-json", wrapFn(fromJSON)},

		&builtinFn{"typeof", typeof},

		&builtinFn{"failure", wrapFn(failure)},
		&builtinFn{"return", wrapFn(returnFn)},
		&builtinFn{"break", wrapFn(breakFn)},
		&builtinFn{"continue", wrapFn(continueFn)},

		&builtinFn{"each", wrapFn(each)},

		&builtinFn{"cd", cd},
		&builtinFn{"visited-dirs", wrapFn(visistedDirs)},
		&builtinFn{"jump-dir", wrapFn(jumpDir)},

		&builtinFn{"source", wrapFn(source)},

		&builtinFn{"+", wrapFn(plus)},
		&builtinFn{"-", wrapFn(minus)},
		&builtinFn{"*", wrapFn(times)},
		&builtinFn{"/", wrapFn(divide)},

		&builtinFn{"=", eq},
	}
}

var (
	argsError  = newFailure("args error")
	inputError = newFailure("input error")
)

var (
	evalCtxType = reflect.TypeOf((*evalCtx)(nil))
	exitusType_ = reflect.TypeOf(exitus{})
	valueType   = reflect.TypeOf((*Value)(nil)).Elem()
)

// wrapFn wraps an inner function into one suitable as a builtin function. It
// generates argument checking and conversion code according to the signature
// of the inner function. The inner function must accept evalCtx* as the first
// argument and return an exitus.
func wrapFn(inner interface{}) func(*evalCtx, []Value) exitus {
	type_ := reflect.TypeOf(inner)
	if type_.In(0) != evalCtxType || type_.Out(0) != exitusType_ {
		panic("bad func")
	}

	requiredArgs := type_.NumIn() - 1
	isVariadic := type_.IsVariadic()
	var variadicType reflect.Type
	if isVariadic {
		requiredArgs -= 1
		variadicType = type_.In(type_.NumIn() - 1).Elem()
		if !supportedIn(variadicType) {
			panic("bad func argument")
		}
	}

	for i := 0; i < requiredArgs; i++ {
		if !supportedIn(type_.In(i + 1)) {
			panic("bad func argument")
		}
	}

	return func(ec *evalCtx, args []Value) exitus {
		if len(args) < requiredArgs || (!isVariadic && len(args) > requiredArgs) {
			return argsError
		}
		callArgs := make([]reflect.Value, len(args)+1)
		callArgs[0] = reflect.ValueOf(ec)

		ok := convertArgs(args[:requiredArgs], callArgs[1:],
			func(i int) reflect.Type { return type_.In(i + 1) })
		if !ok {
			return argsError
		}
		if isVariadic {
			ok := convertArgs(args[requiredArgs:], callArgs[1+requiredArgs:],
				func(i int) reflect.Type { return variadicType })
			if !ok {
				return argsError
			}
		}
		return reflect.ValueOf(inner).Call(callArgs)[0].Interface().(exitus)
	}
}

func supportedIn(t reflect.Type) bool {
	return t.Kind() == reflect.String || t.Kind() == reflect.Float64 ||
		t.Implements(valueType)
}

func convertArgs(args []Value, callArgs []reflect.Value, callType func(int) reflect.Type) bool {
	for i, arg := range args {
		var callArg interface{}
		switch callType(i).Kind() {
		case reflect.String:
			callArg = toString(arg)
		case reflect.Float64:
			var err error
			callArg, err = toFloat(arg)
			if err != nil {
				return false
				// return err
			}
		default:
			if reflect.TypeOf(arg).ConvertibleTo(callType(i)) {
				callArg = arg
			} else {
				return false
				// return argsError
			}
		}
		callArgs[i] = reflect.ValueOf(callArg)
	}
	return true
}

func nop(ec *evalCtx, args []Value) exitus {
	return ok
}

func put(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
	for _, a := range args {
		out <- a
	}
	return ok
}

func typeof(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
	for _, a := range args {
		out <- str(a.Type().String())
	}
	return ok
}

func failure(ec *evalCtx, arg Value) exitus {
	out := ec.ports[1].ch
	out <- newFailure(toString(arg))
	return ok
}

func returnFn(ec *evalCtx) exitus {
	return newFlowExitus(Return)
}

func breakFn(ec *evalCtx) exitus {
	return newFlowExitus(Break)
}

func continueFn(ec *evalCtx) exitus {
	return newFlowExitus(Continue)
}

func print(ec *evalCtx, args ...string) exitus {
	out := ec.ports[1].f
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(arg)
	}
	return ok
}

func println(ec *evalCtx, args ...string) exitus {
	print(ec, args...)
	ec.ports[1].f.WriteString("\n")
	return ok
}

func intoLines(ec *evalCtx) exitus {
	in := ec.ports[0].ch
	out := ec.ports[1].f

	for v := range in {
		fmt.Fprintln(out, toString(v))
	}
	return ok
}

func fromLines(ec *evalCtx) exitus {
	in := ec.ports[0].f
	out := ec.ports[1].ch

	bufferedIn := bufio.NewReader(in)
	for {
		line, err := bufferedIn.ReadString('\n')
		if err == io.EOF {
			return ok
		} else if err != nil {
			return newFailure(err.Error())
		}
		out <- str(line[:len(line)-1])
	}
}

func ratFn(ec *evalCtx, arg Value) exitus {
	out := ec.ports[1].ch
	r, err := toRat(arg)
	if err != nil {
		return newFailure(err.Error())
	}
	out <- r
	return ok
}

// unpack takes any number of tables and output their list elements.
func unpack(ec *evalCtx) exitus {
	in := ec.ports[0].ch
	out := ec.ports[1].ch

	for v := range in {
		if t, ok := v.(*table); !ok {
			return inputError
		} else {
			for _, e := range t.List {
				out <- e
			}
		}
	}

	return ok
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(ec *evalCtx) exitus {
	in := ec.ports[0].f
	out := ec.ports[1].ch

	dec := json.NewDecoder(in)
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return ok
			}
			return newFailure(err.Error())
		}
		out <- fromJSONInterface(v)
	}
}

// each takes a single closure and applies it to all input values.
func each(ec *evalCtx, f *closure) exitus {
	in := ec.ports[0].ch
in:
	for v := range in {
		su := f.Exec(ec.copy("closure of each"), []Value{v})
		// F.Exec will put exactly one stateUpdate on the channel
		e := (<-su).Exitus
		switch e.Sort {
		case Ok, Continue:
			// nop
		case Break:
			break in
		default:
			// TODO wrap it
			return e
		}
	}
	return ok
}

func cd(ec *evalCtx, args []Value) exitus {
	var dir string
	if len(args) == 0 {
		user, err := user.Current()
		if err == nil {
			dir = user.HomeDir
		}
	} else if len(args) == 1 {
		dir = toString(args[0])
	} else {
		return argsError
	}
	err := os.Chdir(dir)
	if err != nil {
		return newFailure(err.Error())
	}
	if ec.store != nil {
		pwd, err := os.Getwd()
		// BUG(xiaq): Possible error of os.Getwd after cd-ing is ignored.
		if err == nil {
			ec.store.AddDir(pwd)
		}
	}
	return ok
}

var storeNotConnected = newFailure("store not connected")

func visistedDirs(ec *evalCtx) exitus {
	if ec.store == nil {
		return storeNotConnected
	}
	dirs, err := ec.store.ListDirs()
	if err != nil {
		return newFailure("store error: " + err.Error())
	}
	out := ec.ports[1].ch
	for _, dir := range dirs {
		table := newTable()
		table.Dict["path"] = str(dir.Path)
		table.Dict["score"] = str(fmt.Sprint(dir.Score))
		out <- table
	}
	return ok
}

var noMatchingDir = newFailure("no matching directory")

func jumpDir(ec *evalCtx, arg string) exitus {
	if ec.store == nil {
		return storeNotConnected
	}
	dirs, err := ec.store.FindDirs(arg)
	if err != nil {
		return newFailure("store error: " + err.Error())
	}
	if len(dirs) == 0 {
		return noMatchingDir
	}
	dir := dirs[0].Path
	err = os.Chdir(dir)
	// TODO(xiaq): Remove directories that no longer exist
	if err != nil {
		return newFailure(err.Error())
	}
	ec.store.AddDir(dir)
	return ok
}

func source(ec *evalCtx, fname string) exitus {
	ec.Source(fname)
	return ok
}

func toFloat(arg Value) (float64, error) {
	arg, ok := arg.(str)
	if !ok {
		return 0, fmt.Errorf("must be string")
	}
	num, err := strconv.ParseFloat(string(arg.(str)), 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func plus(ec *evalCtx, nums ...float64) exitus {
	out := ec.ports[1].ch
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- str(fmt.Sprintf("%g", sum))
	return ok
}

func minus(ec *evalCtx, sum float64, nums ...float64) exitus {
	out := ec.ports[1].ch
	for _, f := range nums {
		sum -= f
	}
	out <- str(fmt.Sprintf("%g", sum))
	return ok
}

func times(ec *evalCtx, nums ...float64) exitus {
	out := ec.ports[1].ch
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- str(fmt.Sprintf("%g", prod))
	return ok
}

func divide(ec *evalCtx, prod float64, nums ...float64) exitus {
	out := ec.ports[1].ch
	for _, f := range nums {
		prod /= f
	}
	out <- str(fmt.Sprintf("%g", prod))
	return ok
}

func eq(ec *evalCtx, args []Value) exitus {
	out := ec.ports[1].ch
	if len(args) == 0 {
		return argsError
	}
	for i := 0; i+1 < len(args); i++ {
		if !valueEq(args[i], args[i+1]) {
			out <- boolean(false)
			return ok
		}
	}
	out <- boolean(true)
	return ok
}
