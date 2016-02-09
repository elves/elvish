package eval

// Builtin functions.

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

var builtinFns []*BuiltinFn

func init() {
	// Needed to work around init loop.
	builtinFns = []*BuiltinFn{
		&BuiltinFn{":", nop},
		&BuiltinFn{"true", nop},

		&BuiltinFn{"print", wrapFn(print)},
		&BuiltinFn{"println", wrapFn(println)},

		&BuiltinFn{"into-lines", wrapFn(intoLines)},
		&BuiltinFn{"from-lines", wrapFn(fromLines)},

		&BuiltinFn{"rat", wrapFn(ratFn)},

		&BuiltinFn{"put", put},
		&BuiltinFn{"put-all", wrapFn(putAll)},
		&BuiltinFn{"unpack", wrapFn(unpack)},

		&BuiltinFn{"from-json", wrapFn(fromJSON)},

		&BuiltinFn{"typeof", typeof},

		&BuiltinFn{"fail", wrapFn(fail)},
		&BuiltinFn{"multi-error", wrapFn(multiErrorFn)},
		&BuiltinFn{"return", wrapFn(returnFn)},
		&BuiltinFn{"break", wrapFn(breakFn)},
		&BuiltinFn{"continue", wrapFn(continueFn)},

		&BuiltinFn{"each", wrapFn(each)},

		&BuiltinFn{"cd", cd},
		&BuiltinFn{"dirs", wrapFn(dirs)},
		&BuiltinFn{"jump", wrapFn(jump)},

		&BuiltinFn{"source", wrapFn(source)},

		&BuiltinFn{"+", wrapFn(plus)},
		&BuiltinFn{"-", wrapFn(minus)},
		&BuiltinFn{"mul", wrapFn(times)},
		&BuiltinFn{"div", wrapFn(divide)},

		&BuiltinFn{"=", eq},
		&BuiltinFn{"deepeq", deepeq},

		&BuiltinFn{"take", wrapFn(take)},
		&BuiltinFn{"drop", wrapFn(drop)},

		&BuiltinFn{"bind", wrapFn(bind)},
		&BuiltinFn{"le", wrapFn(le)},

		&BuiltinFn{"-sleep", wrapFn(_sleep)},
		&BuiltinFn{"-stack", wrapFn(_stack)},
	}
}

var (
	ErrArgs              = errors.New("args error")
	ErrInput             = errors.New("input error")
	ErrStoreNotConnected = errors.New("store not connected")
	ErrNoMatchingDir     = errors.New("no matching directory")
	ErrNoEditor          = errors.New("no line editor")
)

var (
	evalCtxType = reflect.TypeOf((*evalCtx)(nil))
	valueType   = reflect.TypeOf((*Value)(nil)).Elem()
)

// wrapFn wraps an inner function into one suitable as a builtin function. It
// generates argument checking and conversion code according to the signature
// of the inner function. The inner function must accept evalCtx* as the first
// argument and return an exitus.
func wrapFn(inner interface{}) func(*evalCtx, []Value) {
	type_ := reflect.TypeOf(inner)
	if type_.In(0) != evalCtxType {
		panic("bad func")
	}

	requiredArgs := type_.NumIn() - 1
	isVariadic := type_.IsVariadic()
	var variadicType reflect.Type
	if isVariadic {
		requiredArgs--
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

	return func(ec *evalCtx, args []Value) {
		if len(args) < requiredArgs || (!isVariadic && len(args) > requiredArgs) {
			throw(ErrArgs)
		}
		callArgs := make([]reflect.Value, len(args)+1)
		callArgs[0] = reflect.ValueOf(ec)

		ok := convertArgs(args[:requiredArgs], callArgs[1:],
			func(i int) reflect.Type { return type_.In(i + 1) })
		if !ok {
			throw(ErrArgs)
		}
		if isVariadic {
			ok := convertArgs(args[requiredArgs:], callArgs[1+requiredArgs:],
				func(i int) reflect.Type { return variadicType })
			if !ok {
				throw(ErrArgs)
			}
		}
		reflect.ValueOf(inner).Call(callArgs)
	}
}

func supportedIn(t reflect.Type) bool {
	return t.Kind() == reflect.String ||
		t.Kind() == reflect.Int || t.Kind() == reflect.Float64 ||
		t.Implements(valueType)
}

func convertArgs(args []Value, callArgs []reflect.Value, callType func(int) reflect.Type) bool {
	for i, arg := range args {
		var callArg interface{}
		switch callType(i).Kind() {
		case reflect.String:
			callArg = ToString(arg)
		case reflect.Int:
			var err error
			callArg, err = toInt(arg)
			if err != nil {
				return false
			}
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

func nop(ec *evalCtx, args []Value) {
}

func put(ec *evalCtx, args []Value) {
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

func putAll(ec *evalCtx, lists ...List) {
	out := ec.ports[1].Chan
	for _, list := range lists {
		for _, x := range *list.inner {
			out <- x
		}
	}
}

func typeof(ec *evalCtx, args []Value) {
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- String(a.Type().String())
	}
}

func fail(ec *evalCtx, arg Value) {
	throw(errors.New(ToString(arg)))
}

func multiErrorFn(ec *evalCtx, args ...Error) {
	throw(multiError{args})
}

func returnFn(ec *evalCtx) {
	throw(Return)
}

func breakFn(ec *evalCtx) {
	throw(Break)
}

func continueFn(ec *evalCtx) {
	throw(Continue)
}

func print(ec *evalCtx, args ...string) {
	out := ec.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(arg)
	}
}

func println(ec *evalCtx, args ...string) {
	print(ec, args...)
	ec.ports[1].File.WriteString("\n")
}

func intoLines(ec *evalCtx) {
	in := ec.ports[0].Chan
	out := ec.ports[1].File

	for v := range in {
		fmt.Fprintln(out, ToString(v))
	}
}

func fromLines(ec *evalCtx) {
	in := ec.ports[0].File
	out := ec.ports[1].Chan

	bufferedIn := bufio.NewReader(in)
	for {
		line, err := bufferedIn.ReadString('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			throw(err)
		}
		out <- String(line[:len(line)-1])
	}
}

func ratFn(ec *evalCtx, arg Value) {
	out := ec.ports[1].Chan
	r, err := ToRat(arg)
	if err != nil {
		throw(err)
	}
	out <- r
}

// unpack takes any number of tables and output their list elements.
func unpack(ec *evalCtx) {
	in := ec.ports[0].Chan
	out := ec.ports[1].Chan

	for v := range in {
		if list, ok := v.(List); !ok {
			throw(ErrInput)
		} else {
			for _, e := range *list.inner {
				out <- e
			}
		}
	}
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(ec *evalCtx) {
	in := ec.ports[0].File
	out := ec.ports[1].Chan

	dec := json.NewDecoder(in)
	var v interface{}
	for {
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				return
			}
			throw(err)
		}
		out <- FromJSONInterface(v)
	}
}

// each takes a single closure and applies it to all input values.
func each(ec *evalCtx, f *Closure) {
	in := ec.ports[0].Chan
in:
	for v := range in {
		newec := ec.fork("closure of each")
		ex := newec.PCall(f, []Value{v})
		ClosePorts(newec.ports)

		switch ex {
		case nil, Continue:
			// nop
		case Break:
			break in
		default:
			throw(ex)
		}
	}
}

func cd(ec *evalCtx, args []Value) {
	var dir string
	if len(args) == 0 {
		dir = mustGetHome("")
	} else if len(args) == 1 {
		dir = ToString(args[0])
	} else {
		throw(ErrArgs)
	}

	cdInner(dir, ec)
}

func cdInner(dir string, ec *evalCtx) {
	err := os.Chdir(dir)
	if err != nil {
		throw(err)
	}
	if ec.store != nil {
		pwd, err := os.Getwd()
		// BUG(xiaq): Possible error of os.Getwd after cd-ing is ignored.
		if err == nil {
			ec.store.AddDir(pwd)
		}
	}
}

func dirs(ec *evalCtx) {
	if ec.store == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.store.ListDirs()
	if err != nil {
		throw(errors.New("store error: " + err.Error()))
	}
	out := ec.ports[1].Chan
	for _, dir := range dirs {
		out <- Map{&map[Value]Value{
			String("path"):  String(dir.Path),
			String("score"): String(fmt.Sprint(dir.Score)),
		}}
	}
}

func jump(ec *evalCtx, arg string) {
	if ec.store == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.store.FindDirs(arg)
	if err != nil {
		throw(errors.New("store error: " + err.Error()))
	}
	if len(dirs) == 0 {
		throw(ErrNoMatchingDir)
	}
	dir := dirs[0].Path
	err = os.Chdir(dir)
	// TODO(xiaq): Remove directories that no longer exist
	if err != nil {
		throw(err)
	}
	ec.store.AddDir(dir)
}

func source(ec *evalCtx, fname string) {
	ec.Source(fname)
}

func toFloat(arg Value) (float64, error) {
	arg, ok := arg.(String)
	if !ok {
		return 0, fmt.Errorf("must be string")
	}
	num, err := strconv.ParseFloat(string(arg.(String)), 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func toInt(arg Value) (int, error) {
	arg, ok := arg.(String)
	if !ok {
		return 0, fmt.Errorf("must be string")
	}
	num, err := strconv.Atoi(string(arg.(String)))
	if err != nil {
		return 0, err
	}
	return num, nil
}

func plus(ec *evalCtx, nums ...float64) {
	out := ec.ports[1].Chan
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- String(fmt.Sprintf("%g", sum))
}

func minus(ec *evalCtx, sum float64, nums ...float64) {
	out := ec.ports[1].Chan
	for _, f := range nums {
		sum -= f
	}
	out <- String(fmt.Sprintf("%g", sum))
}

func times(ec *evalCtx, nums ...float64) {
	out := ec.ports[1].Chan
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- String(fmt.Sprintf("%g", prod))
}

func divide(ec *evalCtx, prod float64, nums ...float64) {
	out := ec.ports[1].Chan
	for _, f := range nums {
		prod /= f
	}
	out <- String(fmt.Sprintf("%g", prod))
}

func eq(ec *evalCtx, args []Value) {
	out := ec.ports[1].Chan
	if len(args) == 0 {
		throw(ErrArgs)
	}
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			out <- Bool(false)
			return
		}
	}
	out <- Bool(true)
}

func deepeq(ec *evalCtx, args []Value) {
	out := ec.ports[1].Chan
	if len(args) == 0 {
		throw(ErrArgs)
	}
	for i := 0; i+1 < len(args); i++ {
		if !DeepEq(args[i], args[i+1]) {
			out <- Bool(false)
			return
		}
	}
	out <- Bool(true)
}

func take(ec *evalCtx, n int) {
	in := ec.ports[0].Chan
	out := ec.ports[1].Chan

	i := 0
	for v := range in {
		if i >= n {
			break
		}
		i++
		out <- v
	}
}

func drop(ec *evalCtx, n int) {
	in := ec.ports[0].Chan
	out := ec.ports[1].Chan

	for i := 0; i < n; i++ {
		<-in
	}
	for v := range in {
		out <- v
	}
}

func bind(ec *evalCtx, key string, function Value) {
	if ec.Editor == nil {
		throw(ErrNoEditor)
	}
	maybeThrow(ec.Editor.Bind(key, function))
}

func le(ec *evalCtx, name string, args ...Value) {
	if ec.Editor == nil {
		throw(ErrNoEditor)
	}
	maybeThrow(ec.Editor.Call(name, args))
}

func _sleep(ec *evalCtx, t float64) {
	time.Sleep(time.Duration(t) * time.Second)
}

func _stack(ec *evalCtx) {
	out := ec.ports[1].File

	// XXX dup with main.go
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)
}
