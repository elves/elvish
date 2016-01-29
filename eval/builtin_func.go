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
	"runtime"
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
		&builtinFn{"put-all", wrapFn(putAll)},
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

		&builtinFn{"ele", wrapFn(ele)},

		&builtinFn{"-stack", wrapFn(_stack)},
	}
}

var (
	argsError  = NewFailure("args error")
	inputError = NewFailure("input error")
)

var (
	evalCtxType = reflect.TypeOf((*evalCtx)(nil))
	exitusType_ = reflect.TypeOf(Exitus{})
	valueType   = reflect.TypeOf((*Value)(nil)).Elem()
)

// wrapFn wraps an inner function into one suitable as a builtin function. It
// generates argument checking and conversion code according to the signature
// of the inner function. The inner function must accept evalCtx* as the first
// argument and return an exitus.
func wrapFn(inner interface{}) func(*evalCtx, []Value) Exitus {
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

	return func(ec *evalCtx, args []Value) Exitus {
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
		return reflect.ValueOf(inner).Call(callArgs)[0].Interface().(Exitus)
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

func nop(ec *evalCtx, args []Value) Exitus {
	return OK
}

func put(ec *evalCtx, args []Value) Exitus {
	out := ec.ports[1].ch
	for _, a := range args {
		out <- a
	}
	return OK
}

func putAll(ec *evalCtx, lists ...*list) Exitus {
	out := ec.ports[1].ch
	for _, list := range lists {
		for _, x := range *list {
			out <- x
		}
	}
	return OK
}

func typeof(ec *evalCtx, args []Value) Exitus {
	out := ec.ports[1].ch
	for _, a := range args {
		out <- str(a.Type().String())
	}
	return OK
}

func failure(ec *evalCtx, arg Value) Exitus {
	out := ec.ports[1].ch
	out <- NewFailure(toString(arg))
	return OK
}

func returnFn(ec *evalCtx) Exitus {
	return newFlowExitus(Return)
}

func breakFn(ec *evalCtx) Exitus {
	return newFlowExitus(Break)
}

func continueFn(ec *evalCtx) Exitus {
	return newFlowExitus(Continue)
}

func print(ec *evalCtx, args ...string) Exitus {
	out := ec.ports[1].f
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(arg)
	}
	return OK
}

func println(ec *evalCtx, args ...string) Exitus {
	print(ec, args...)
	ec.ports[1].f.WriteString("\n")
	return OK
}

func intoLines(ec *evalCtx) Exitus {
	in := ec.ports[0].ch
	out := ec.ports[1].f

	for v := range in {
		fmt.Fprintln(out, toString(v))
	}
	return OK
}

func fromLines(ec *evalCtx) Exitus {
	in := ec.ports[0].f
	out := ec.ports[1].ch

	bufferedIn := bufio.NewReader(in)
	for {
		line, err := bufferedIn.ReadString('\n')
		if err == io.EOF {
			return OK
		} else if err != nil {
			return NewFailure(err.Error())
		}
		out <- str(line[:len(line)-1])
	}
}

func ratFn(ec *evalCtx, arg Value) Exitus {
	out := ec.ports[1].ch
	r, err := toRat(arg)
	if err != nil {
		return NewFailure(err.Error())
	}
	out <- r
	return OK
}

// unpack takes any number of tables and output their list elements.
func unpack(ec *evalCtx) Exitus {
	in := ec.ports[0].ch
	out := ec.ports[1].ch

	for v := range in {
		if list, ok := v.(*list); !ok {
			return inputError
		} else {
			for _, e := range *list {
				out <- e
			}
		}
	}

	return OK
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(ec *evalCtx) Exitus {
	in := ec.ports[0].f
	out := ec.ports[1].ch

	dec := json.NewDecoder(in)
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return OK
			}
			return NewFailure(err.Error())
		}
		out <- fromJSONInterface(v)
	}
}

// each takes a single closure and applies it to all input values.
func each(ec *evalCtx, f *closure) Exitus {
	in := ec.ports[0].ch
in:
	for v := range in {
		newec := ec.fork("closure of each")
		ex := f.Call(newec, []Value{v})
		newec.closePorts()

		switch ex.Sort {
		case Ok, Continue:
			// nop
		case Break:
			break in
		default:
			// TODO wrap it
			return ex
		}
	}
	return OK
}

func cd(ec *evalCtx, args []Value) Exitus {
	var dir string
	if len(args) == 0 {
		user, err := user.Current()
		if err == nil {
			dir = user.HomeDir
		} else {
			return NewFailure("cannot get current user: " + err.Error())
		}
	} else if len(args) == 1 {
		dir = toString(args[0])
	} else {
		return argsError
	}

	return cdInner(dir, ec)
}

func cdInner(dir string, ec *evalCtx) Exitus {
	err := os.Chdir(dir)
	if err != nil {
		return NewFailure(err.Error())
	}
	if ec.store != nil {
		pwd, err := os.Getwd()
		// BUG(xiaq): Possible error of os.Getwd after cd-ing is ignored.
		if err == nil {
			ec.store.AddDir(pwd)
		}
	}
	return OK
}

var storeNotConnected = NewFailure("store not connected")

func visistedDirs(ec *evalCtx) Exitus {
	if ec.store == nil {
		return storeNotConnected
	}
	dirs, err := ec.store.ListDirs()
	if err != nil {
		return NewFailure("store error: " + err.Error())
	}
	out := ec.ports[1].ch
	for _, dir := range dirs {
		m := newMap()
		m["path"] = str(dir.Path)
		m["score"] = str(fmt.Sprint(dir.Score))
		out <- m
	}
	return OK
}

var noMatchingDir = NewFailure("no matching directory")

func jumpDir(ec *evalCtx, arg string) Exitus {
	if ec.store == nil {
		return storeNotConnected
	}
	dirs, err := ec.store.FindDirs(arg)
	if err != nil {
		return NewFailure("store error: " + err.Error())
	}
	if len(dirs) == 0 {
		return noMatchingDir
	}
	dir := dirs[0].Path
	err = os.Chdir(dir)
	// TODO(xiaq): Remove directories that no longer exist
	if err != nil {
		return NewFailure(err.Error())
	}
	ec.store.AddDir(dir)
	return OK
}

func source(ec *evalCtx, fname string) Exitus {
	ec.Source(fname)
	return OK
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

func plus(ec *evalCtx, nums ...float64) Exitus {
	out := ec.ports[1].ch
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- str(fmt.Sprintf("%g", sum))
	return OK
}

func minus(ec *evalCtx, sum float64, nums ...float64) Exitus {
	out := ec.ports[1].ch
	for _, f := range nums {
		sum -= f
	}
	out <- str(fmt.Sprintf("%g", sum))
	return OK
}

func times(ec *evalCtx, nums ...float64) Exitus {
	out := ec.ports[1].ch
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- str(fmt.Sprintf("%g", prod))
	return OK
}

func divide(ec *evalCtx, prod float64, nums ...float64) Exitus {
	out := ec.ports[1].ch
	for _, f := range nums {
		prod /= f
	}
	out <- str(fmt.Sprintf("%g", prod))
	return OK
}

func eq(ec *evalCtx, args []Value) Exitus {
	out := ec.ports[1].ch
	if len(args) == 0 {
		return argsError
	}
	for i := 0; i+1 < len(args); i++ {
		if !valueEq(args[i], args[i+1]) {
			out <- boolean(false)
			return OK
		}
	}
	out <- boolean(true)
	return OK
}

var noEditor = NewFailure("no line editor")

func ele(ec *evalCtx, name string, args ...Value) Exitus {
	if ec.Editor == nil {
		return noEditor
	}
	ec.Editor.Call(name, args)
	return OK
}

func _stack(ec *evalCtx) Exitus {
	out := ec.ports[1].f

	// XXX dup with main.go
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)

	return OK
}
