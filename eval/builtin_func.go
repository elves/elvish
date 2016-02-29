package eval

// Builtin functions.

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var builtinFns []*BuiltinFn

// BuiltinFn is a builtin function.
type BuiltinFn struct {
	Name string
	Impl func(*EvalCtx, []Value)
}

func (*BuiltinFn) Kind() string {
	return "fn"
}

func (b *BuiltinFn) Repr(int) string {
	return "$" + FnPrefix + b.Name
}

// Call calls a builtin function.
func (b *BuiltinFn) Call(ec *EvalCtx, args []Value) {
	b.Impl(ec, args)
}

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

		&BuiltinFn{"kind-of", kindOf},

		&BuiltinFn{"fail", wrapFn(fail)},
		&BuiltinFn{"multi-error", wrapFn(multiErrorFn)},
		&BuiltinFn{"return", wrapFn(returnFn)},
		&BuiltinFn{"break", wrapFn(breakFn)},
		&BuiltinFn{"continue", wrapFn(continueFn)},

		&BuiltinFn{"each", wrapFn(each)},

		&BuiltinFn{"cd", cd},
		&BuiltinFn{"dirs", wrapFn(dirs)},
		&BuiltinFn{"history", wrapFn(history)},

		&BuiltinFn{"source", wrapFn(source)},

		&BuiltinFn{"+", wrapFn(plus)},
		&BuiltinFn{"-", wrapFn(minus)},
		&BuiltinFn{"mul", wrapFn(times)},
		&BuiltinFn{"div", wrapFn(divide)},
		&BuiltinFn{"pow", wrapFn(pow)},
		&BuiltinFn{"lt", wrapFn(lt)},
		&BuiltinFn{"gt", wrapFn(gt)},

		&BuiltinFn{"base", wrapFn(base)},

		&BuiltinFn{"=", eq},
		&BuiltinFn{"deepeq", deepeq},

		&BuiltinFn{"take", wrapFn(take)},
		&BuiltinFn{"drop", wrapFn(drop)},

		&BuiltinFn{"len", wrapFn(lenFn)},
		&BuiltinFn{"count", wrapFn(count)},

		&BuiltinFn{"fg", wrapFn(fg)},

		&BuiltinFn{"tilde-abbr", wrapFn(tildeAbbr)},

		&BuiltinFn{"-sleep", wrapFn(_sleep)},
		&BuiltinFn{"-stack", wrapFn(_stack)},
		&BuiltinFn{"-log", wrapFn(_log)},
		&BuiltinFn{"-exec", wrapFn(_exec)},
	}
	for _, b := range builtinFns {
		builtinNamespace[FnPrefix+b.Name] = NewRoVariable(b)
	}
}

var (
	ErrArgs              = errors.New("args error")
	ErrInput             = errors.New("input error")
	ErrStoreNotConnected = errors.New("store not connected")
	ErrNoMatchingDir     = errors.New("no matching directory")
	ErrNotInSameGroup    = errors.New("not in the same process group")
	ErrInterrupted       = errors.New("interrupted")
)

var (
	evalCtxType = reflect.TypeOf((*EvalCtx)(nil))
	valueType   = reflect.TypeOf((*Value)(nil)).Elem()
)

// wrapFn wraps an inner function into one suitable as a builtin function. It
// generates argument checking and conversion code according to the signature
// of the inner function. The inner function must accept evalCtx* as the first
// argument and return an exitus.
func wrapFn(inner interface{}) func(*EvalCtx, []Value) {
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

	return func(ec *EvalCtx, args []Value) {
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

func nop(ec *EvalCtx, args []Value) {
}

func put(ec *EvalCtx, args []Value) {
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

func putAll(ec *EvalCtx, lists ...List) {
	out := ec.ports[1].Chan
	for _, list := range lists {
		for _, x := range *list.inner {
			out <- x
		}
	}
}

func kindOf(ec *EvalCtx, args []Value) {
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- String(a.Kind())
	}
}

func fail(ec *EvalCtx, arg Value) {
	throw(errors.New(ToString(arg)))
}

func multiErrorFn(ec *EvalCtx, args ...Error) {
	throw(MultiError{args})
}

func returnFn(ec *EvalCtx) {
	throw(Return)
}

func breakFn(ec *EvalCtx) {
	throw(Break)
}

func continueFn(ec *EvalCtx) {
	throw(Continue)
}

func print(ec *EvalCtx, args ...string) {
	out := ec.ports[1].File
	for i, arg := range args {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(arg)
	}
}

func println(ec *EvalCtx, args ...string) {
	print(ec, args...)
	ec.ports[1].File.WriteString("\n")
}

func intoLines(ec *EvalCtx) {
	in := ec.ports[0].Chan
	out := ec.ports[1].File

	for v := range in {
		fmt.Fprintln(out, ToString(v))
	}
}

func fromLines(ec *EvalCtx) {
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

func ratFn(ec *EvalCtx, arg Value) {
	out := ec.ports[1].Chan
	r, err := ToRat(arg)
	if err != nil {
		throw(err)
	}
	out <- r
}

// unpack takes Elemser's from the input and unpack them.
func unpack(ec *EvalCtx) {
	in := ec.ports[0].Chan
	out := ec.ports[1].Chan

	for v := range in {
		elemser, ok := v.(Elemser)
		if !ok {
			throw(ErrInput)
		}
		for e := range elemser.Elems() {
			out <- e
		}
	}
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(ec *EvalCtx) {
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
func each(ec *EvalCtx, f CallerValue) {
	in := ec.ports[0].Chan
in:
	for v := range in {
		// NOTE We don't have the position range of the closure in the source.
		// Ideally, it should be kept in the Closure itself.
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

func cd(ec *EvalCtx, args []Value) {
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

func cdInner(dir string, ec *EvalCtx) {
	err := os.Chdir(dir)
	if err != nil {
		throw(err)
	}
	if ec.store != nil {
		// XXX Error ignored.
		pwd, err := os.Getwd()
		if err == nil {
			store := ec.store
			go func() {
				store.Waits.Add(1)
				// XXX Error ignored.
				store.AddDir(pwd, 1)
				store.Waits.Done()
				Logger.Println("added dir to store:", pwd)
			}()
		}
	}
}

var dirFieldNames = []string{"path", "score"}

func dirs(ec *EvalCtx) {
	if ec.store == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.store.ListDirs()
	if err != nil {
		throw(errors.New("store error: " + err.Error()))
	}
	out := ec.ports[1].Chan
	for _, dir := range dirs {
		out <- &Struct{dirFieldNames, []Variable{
			NewRoVariable(String(dir.Path)),
			NewRoVariable(String(fmt.Sprint(dir.Score))),
		}}
	}
}

func history(ec *EvalCtx) {
	if ec.store == nil {
		throw(ErrStoreNotConnected)
	}

	store := ec.store
	seq, err := store.NextCmdSeq()
	maybeThrow(err)
	cmds, err := store.Cmds(0, seq)
	maybeThrow(err)

	out := ec.ports[1].Chan
	for _, cmd := range cmds {
		out <- String(cmd)
	}
}

func source(ec *EvalCtx, fname string) {
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

func plus(ec *EvalCtx, nums ...float64) {
	out := ec.ports[1].Chan
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- String(fmt.Sprintf("%g", sum))
}

func minus(ec *EvalCtx, sum float64, nums ...float64) {
	out := ec.ports[1].Chan
	for _, f := range nums {
		sum -= f
	}
	out <- String(fmt.Sprintf("%g", sum))
}

func times(ec *EvalCtx, nums ...float64) {
	out := ec.ports[1].Chan
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	out <- String(fmt.Sprintf("%g", prod))
}

func divide(ec *EvalCtx, prod float64, nums ...float64) {
	out := ec.ports[1].Chan
	for _, f := range nums {
		prod /= f
	}
	out <- String(fmt.Sprintf("%g", prod))
}

func pow(ec *EvalCtx, b, p float64) {
	out := ec.ports[1].Chan
	out <- String(fmt.Sprintf("%g", math.Pow(b, p)))
}

var ErrFalse = errors.New("false")

func lt(ec *EvalCtx, nums ...float64) {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] < nums[i+1]) {
			throw(ErrFalse)
		}
	}
}

func gt(ec *EvalCtx, nums ...float64) {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] > nums[i+1]) {
			throw(ErrFalse)
		}
	}
}

var ErrBadBase = errors.New("bad base")

func base(ec *EvalCtx, b int, nums ...int) {
	if b < 2 || b > 36 {
		throw(ErrBadBase)
	}

	out := ec.ports[1].Chan

	for _, num := range nums {
		out <- String(strconv.FormatInt(int64(num), b))
	}
}

var ErrNotEqual = errors.New("not equal")

func eq(ec *EvalCtx, args []Value) {
	if len(args) == 0 {
		throw(ErrArgs)
	}
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			throw(ErrNotEqual)
		}
	}
}

func deepeq(ec *EvalCtx, args []Value) {
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

func take(ec *EvalCtx, n int) {
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

func drop(ec *EvalCtx, n int) {
	in := ec.ports[0].Chan
	out := ec.ports[1].Chan

	for i := 0; i < n; i++ {
		<-in
	}
	for v := range in {
		out <- v
	}
}

func lenFn(ec *EvalCtx, v Value) {
	lener, ok := v.(Lener)
	if !ok {
		throw(fmt.Errorf("cannot get length of a %s", v.Kind()))
	}
	ec.ports[1].Chan <- String(strconv.Itoa(lener.Len()))
}

func count(ec *EvalCtx) {
	in := ec.ports[0].Chan
	out := ec.ports[1].Chan

	n := 0
	for range in {
		n++
	}
	out <- String(strconv.Itoa(n))
}

func fg(ec *EvalCtx, pids ...int) {
	if len(pids) == 0 {
		throw(ErrArgs)
	}
	var thepgid int
	for i, pid := range pids {
		pgid, err := syscall.Getpgid(pid)
		maybeThrow(err)
		if i == 0 {
			thepgid = pgid
		} else if pgid != thepgid {
			throw(ErrNotInSameGroup)
		}
	}

	err := sys.Tcsetpgrp(0, thepgid)
	maybeThrow(err)

	errors := make([]Error, len(pids))

	for i, pid := range pids {
		err := syscall.Kill(pid, syscall.SIGCONT)
		if err != nil {
			errors[i] = Error{err}
		}
	}

	for i, pid := range pids {
		if errors[i] != OK {
			continue
		}
		var ws syscall.WaitStatus
		_, err = syscall.Wait4(pid, &ws, syscall.WUNTRACED, nil)
		if err != nil {
			errors[i] = Error{err}
		} else {
			errors[i] = Error{NewExternalCmdExit(ws, pid)}
		}
	}

	throwCompositeError(errors)
}

func tildeAbbr(ec *EvalCtx, path string) {
	out := ec.ports[1].Chan
	out <- String(util.TildeAbbr(path))
}

func _sleep(ec *EvalCtx, t float64) {
	d := time.Duration(float64(time.Second) * t)
	select {
	case <-ec.intCh:
		throw(ErrInterrupted)
	case <-time.After(d):
	}
}

func _stack(ec *EvalCtx) {
	out := ec.ports[1].File

	// XXX dup with main.go
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)
}

func _log(ec *EvalCtx, fname string) {
	maybeThrow(util.SetOutputFile(fname))
}

func _exec(ec *EvalCtx, args ...string) {
	if len(args) == 0 {
		args = []string{"elvish"}
	}
	var err error
	args[0], err = ec.Search(args[0])
	maybeThrow(err)
	err = ec.store.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if ec.Stub != nil {
		ec.Stub.Terminate()
	}
	err = syscall.Exec(args[0], args, os.Environ())
	maybeThrow(err)
}
