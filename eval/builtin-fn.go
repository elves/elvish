package eval

// Builtin functions.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var builtinFns []*BuiltinFn

// BuiltinFn is a builtin function.
type BuiltinFn struct {
	Name string
	Impl func(*EvalCtx, []Value, map[string]Value)
}

var _ FnValue = &BuiltinFn{}

func (*BuiltinFn) Kind() string {
	return "fn"
}

func (b *BuiltinFn) Repr(int) string {
	return "<builtin " + b.Name + ">"
}

// Call calls a builtin function.
func (b *BuiltinFn) Call(ec *EvalCtx, args []Value, opts map[string]Value) {
	b.Impl(ec, args, opts)
}

func init() {
	// Needed to work around init loop.
	builtinFns = []*BuiltinFn{
		&BuiltinFn{"true", nop},
		&BuiltinFn{"false", falseFn},

		&BuiltinFn{"print", WrapFn(print, OptSpec{"sep", String(" ")})},
		&BuiltinFn{"echo", WrapFn(echo, OptSpec{"sep", String(" ")})},
		&BuiltinFn{"pprint", pprint},

		&BuiltinFn{"slurp", WrapFn(slurp)},
		&BuiltinFn{"into-lines", WrapFn(intoLines)},

		&BuiltinFn{"put", put},
		&BuiltinFn{"unpack", WrapFn(unpack)},

		&BuiltinFn{"joins", WrapFn(joins)},
		&BuiltinFn{"splits", WrapFn(splits, OptSpec{"sep", String("")})},
		&BuiltinFn{"has-prefix", WrapFn(hasPrefix)},
		&BuiltinFn{"has-suffix", WrapFn(hasSuffix)},
		&BuiltinFn{"<s",
			wrapStrCompare(func(a, b string) bool { return a < b })},
		&BuiltinFn{"<=s",
			wrapStrCompare(func(a, b string) bool { return a <= b })},
		&BuiltinFn{"==s",
			wrapStrCompare(func(a, b string) bool { return a == b })},
		&BuiltinFn{"!=s",
			wrapStrCompare(func(a, b string) bool { return a != b })},
		&BuiltinFn{">s",
			wrapStrCompare(func(a, b string) bool { return a > b })},
		&BuiltinFn{">=s",
			wrapStrCompare(func(a, b string) bool { return a >= b })},

		&BuiltinFn{"to-json", WrapFn(toJSON)},
		&BuiltinFn{"from-json", WrapFn(fromJSON)},

		&BuiltinFn{"kind-of", kindOf},

		&BuiltinFn{"fail", WrapFn(fail)},
		&BuiltinFn{"multi-error", WrapFn(multiErrorFn)},
		&BuiltinFn{"return", WrapFn(returnFn)},
		&BuiltinFn{"break", WrapFn(breakFn)},
		&BuiltinFn{"continue", WrapFn(continueFn)},

		&BuiltinFn{"each", WrapFn(each)},
		&BuiltinFn{"peach", WrapFn(peach)},
		&BuiltinFn{"eawk", WrapFn(eawk)},
		&BuiltinFn{"constantly", constantly},

		&BuiltinFn{"cd", cd},
		&BuiltinFn{"dirs", WrapFn(dirs)},
		&BuiltinFn{"history", WrapFn(history)},

		&BuiltinFn{"path-abs", wrapStringToStringError(filepath.Abs)},
		&BuiltinFn{"path-base", wrapStringToString(filepath.Base)},
		&BuiltinFn{"path-clean", wrapStringToString(filepath.Clean)},
		&BuiltinFn{"path-dir", wrapStringToString(filepath.Dir)},
		&BuiltinFn{"path-ext", wrapStringToString(filepath.Ext)},
		&BuiltinFn{"eval-symlinks", wrapStringToStringError(filepath.EvalSymlinks)},

		&BuiltinFn{"source", WrapFn(source)},

		&BuiltinFn{"+", WrapFn(plus)},
		&BuiltinFn{"-", WrapFn(minus)},
		&BuiltinFn{"*", WrapFn(times)},
		&BuiltinFn{"/", slash},
		&BuiltinFn{"^", WrapFn(pow)},
		&BuiltinFn{"<",
			wrapNumCompare(func(a, b float64) bool { return a < b })},
		&BuiltinFn{"<=",
			wrapNumCompare(func(a, b float64) bool { return a <= b })},
		&BuiltinFn{"==",
			wrapNumCompare(func(a, b float64) bool { return a == b })},
		&BuiltinFn{"!=",
			wrapNumCompare(func(a, b float64) bool { return a != b })},
		&BuiltinFn{">",
			wrapNumCompare(func(a, b float64) bool { return a > b })},
		&BuiltinFn{">=",
			wrapNumCompare(func(a, b float64) bool { return a >= b })},
		&BuiltinFn{"%", WrapFn(mod)},
		&BuiltinFn{"rand", WrapFn(randFn)},
		&BuiltinFn{"randint", WrapFn(randint)},

		&BuiltinFn{"ord", WrapFn(ord)},
		&BuiltinFn{"base", WrapFn(base)},

		&BuiltinFn{"range", rangeFn},

		&BuiltinFn{"bool", WrapFn(boolFn)},
		&BuiltinFn{"is", is},
		&BuiltinFn{"eq", eq},

		&BuiltinFn{"resolve", WrapFn(resolveFn)},

		&BuiltinFn{"take", WrapFn(take)},

		&BuiltinFn{"count", count},
		&BuiltinFn{"wcswidth", WrapFn(wcswidth)},

		&BuiltinFn{"fg", WrapFn(fg)},

		&BuiltinFn{"tilde-abbr", WrapFn(tildeAbbr)},

		&BuiltinFn{"fopen", WrapFn(fopen)},
		&BuiltinFn{"fclose", WrapFn(fclose)},
		&BuiltinFn{"pipe", WrapFn(pipe)},
		&BuiltinFn{"prclose", WrapFn(prclose)},
		&BuiltinFn{"pwclose", WrapFn(pwclose)},

		&BuiltinFn{"esleep", WrapFn(sleep)},
		&BuiltinFn{"exec", WrapFn(exec)},
		&BuiltinFn{"exit", WrapFn(exit)},

		&BuiltinFn{"-stack", WrapFn(_stack)},
		&BuiltinFn{"-log", WrapFn(_log)},
	}
	for _, b := range builtinFns {
		builtinNamespace[FnPrefix+b.Name] = NewRoVariable(b)
	}

	// For rand and randint.
	rand.Seed(time.Now().UTC().UnixNano())
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
	evalCtxType     = reflect.TypeOf((*EvalCtx)(nil))
	valueType       = reflect.TypeOf((*Value)(nil)).Elem()
	iterateType     = reflect.TypeOf((func(func(Value)))(nil))
	stringValueType = reflect.TypeOf(String(""))
)

// WrapFn wraps an inner function into one suitable as a builtin function. It
// generates argument checking and conversion code according to the signature of
// the inner function and option specifications. The inner function must accept
// EvalCtx* as the first argument, followed by options, followed by arguments.
func WrapFn(inner interface{}, optSpecs ...OptSpec) func(*EvalCtx, []Value, map[string]Value) {
	funcType := reflect.TypeOf(inner)
	if funcType.In(0) != evalCtxType {
		panic("bad func to wrap, first argument not *EvalCtx")
	}

	nopts := len(optSpecs)
	optsTo := nopts + 1
	optSet := NewOptSet(optSpecs...)
	// Range occupied by fixed arguments in the argument list to inner.
	fixedArgsFrom, fixedArgsTo := optsTo, funcType.NumIn()
	isVariadic := funcType.IsVariadic()
	hasOptionalIterate := false
	var variadicType reflect.Type
	if isVariadic {
		fixedArgsTo--
		variadicType = funcType.In(funcType.NumIn() - 1).Elem()
		if !supportedArgType(variadicType) {
			panic(fmt.Sprintf("bad func to wrap, variadic argument type %s unsupported", variadicType))
		}
	} else if funcType.In(funcType.NumIn()-1) == iterateType {
		fixedArgsTo--
		hasOptionalIterate = true
	}

	for i := 1; i < fixedArgsTo; i++ {
		if !supportedArgType(funcType.In(i)) {
			panic(fmt.Sprintf("bad func to wrap, argument type %s unsupported", funcType.In(i)))
		}
	}

	nFixedArgs := fixedArgsTo - fixedArgsFrom

	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		// Check arity of arguments.
		if isVariadic {
			if len(args) < nFixedArgs {
				throw(fmt.Errorf("arity mismatch: want %d or more arguments, got %d", nFixedArgs, len(args)))
			}
		} else if hasOptionalIterate {
			if len(args) < nFixedArgs || len(args) > nFixedArgs+1 {
				throw(fmt.Errorf("arity mismatch: want %d or %d arguments, got %d", nFixedArgs, nFixedArgs+1, len(args)))
			}
		} else if len(args) != nFixedArgs {
			throw(fmt.Errorf("arity mismatch: want %d arguments, got %d", nFixedArgs, len(args)))
		}
		convertedArgs := make([]reflect.Value, 1+nopts+len(args))
		convertedArgs[0] = reflect.ValueOf(ec)

		// Convert and fill options.
		var err error
		optValues := optSet.MustPick(opts)
		for i, v := range optValues {
			convertedArgs[1+i], err = convertArg(v, funcType.In(1+i))
			if err != nil {
				throw(errors.New("bad option " + parse.Quote(optSet.optSpecs[i].Name) + ": " + err.Error()))
			}
		}

		// Convert and fill fixed arguments.
		for i, arg := range args[:nFixedArgs] {
			convertedArgs[fixedArgsFrom+i], err = convertArg(arg, funcType.In(fixedArgsFrom+i))
			if err != nil {
				throw(errors.New("bad argument: " + err.Error()))
			}
		}

		if isVariadic {
			for i, arg := range args[nFixedArgs:] {
				convertedArgs[fixedArgsTo+i], err = convertArg(arg, variadicType)
				if err != nil {
					throw(errors.New("bad argument: " + err.Error()))
				}
			}
		} else if hasOptionalIterate {
			var iterate func(func(Value))
			if len(args) == nFixedArgs {
				// No Iterator specified in arguments. Use input.
				// Since convertedArgs was created according to the size of the
				// actual argument list, we now an empty element to make room
				// for this additional iterator argument.
				convertedArgs = append(convertedArgs, reflect.Value{})
				iterate = ec.IterateInputs
			} else {
				iterator, ok := args[nFixedArgs].(Iterator)
				if !ok {
					throw(errors.New("bad argument: need iterator, got " + args[nFixedArgs].Kind()))
				}
				iterate = func(f func(Value)) {
					iterator.Iterate(func(v Value) bool {
						f(v)
						return true
					})
				}
			}
			convertedArgs[fixedArgsTo] = reflect.ValueOf(iterate)
		}
		reflect.ValueOf(inner).Call(convertedArgs)
	}
}

func supportedArgType(t reflect.Type) bool {
	return t.Kind() == reflect.String ||
		t.Kind() == reflect.Int || t.Kind() == reflect.Float64 ||
		t.Implements(valueType)
}

func convertArg(arg Value, wantType reflect.Type) (reflect.Value, error) {
	var converted interface{}
	var err error

	switch wantType.Kind() {
	case reflect.String:
		if wantType == stringValueType {
			converted = String(ToString(arg))
		} else {
			converted = ToString(arg)
		}
	case reflect.Int:
		converted, err = toInt(arg)
	case reflect.Float64:
		converted, err = toFloat(arg)
	default:
		if reflect.TypeOf(arg).ConvertibleTo(wantType) {
			converted = arg
		} else {
			err = fmt.Errorf("need %s", wantType.Name())
		}
	}
	return reflect.ValueOf(converted), err
}

func wrapStringToString(f func(string) string) func(*EvalCtx, []Value, map[string]Value) {
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		s := mustGetOneString(args)
		ec.ports[1].Chan <- String(f(s))
	}
}

func wrapStringToStringError(f func(string) (string, error)) func(*EvalCtx, []Value, map[string]Value) {
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		s := mustGetOneString(args)
		result, err := f(s)
		maybeThrow(err)
		ec.ports[1].Chan <- String(result)
	}
}

func wrapStrCompare(cmp func(a, b string) bool) func(*EvalCtx, []Value, map[string]Value) {
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		if len(args) < 2 {
			throw(ErrArgs)
		}
		for _, a := range args {
			if _, ok := a.(String); !ok {
				throw(ErrArgs)
			}
		}
		for i := 0; i < len(args)-1; i++ {
			if !cmp(string(args[i].(String)), string(args[i+1].(String))) {
				ec.falsify()
				return
			}
		}
	}
}

func wrapNumCompare(cmp func(a, b float64) bool) func(*EvalCtx, []Value, map[string]Value) {
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		if len(args) < 2 {
			throw(ErrArgs)
		}
		floats := make([]float64, len(args))
		for i, a := range args {
			f, err := toFloat(a)
			maybeThrow(err)
			floats[i] = f
		}
		for i := 0; i < len(floats)-1; i++ {
			if !cmp(floats[i], floats[i+1]) {
				ec.falsify()
				return
			}
		}
	}
}

var errMustBeOneString = errors.New("must be one string argument")

func mustGetOneString(args []Value) string {
	if len(args) != 1 {
		throw(errMustBeOneString)
	}
	s, ok := args[0].(String)
	if !ok {
		throw(errMustBeOneString)
	}
	return string(s)
}

func nop(ec *EvalCtx, args []Value, opts map[string]Value) {
}

func falseFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	ec.falsify()
}

func put(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

func kindOf(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
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

func print(ec *EvalCtx, sepv String, args ...string) {
	out := ec.ports[1].File
	sep := string(sepv)
	for i, arg := range args {
		if i > 0 {
			out.WriteString(sep)
		}
		out.WriteString(arg)
	}
}

func echo(ec *EvalCtx, sep String, args ...string) {
	print(ec, sep, args...)
	ec.ports[1].File.WriteString("\n")
}

func pprint(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	out := ec.ports[1].File
	for _, arg := range args {
		out.WriteString(arg.Repr(0))
		out.WriteString("\n")
	}
}

func slurp(ec *EvalCtx) {
	in := ec.ports[0].File
	out := ec.ports[1].Chan

	all, err := ioutil.ReadAll(in)
	if err != nil {
		b, err := sys.GetNonblock(0)
		fmt.Println("stdin is nonblock:", b, err)
		fmt.Println("stdin is stdin:", in == os.Stdin)
	}
	maybeThrow(err)
	out <- String(string(all))
}

func intoLines(ec *EvalCtx, iterate func(func(Value))) {
	out := ec.ports[1].File

	iterate(func(v Value) {
		fmt.Fprintln(out, ToString(v))
	})
}

// unpack puts each element of the argument.
func unpack(ec *EvalCtx, v IteratorValue) {
	out := ec.ports[1].Chan
	v.Iterate(func(e Value) bool {
		out <- e
		return true
	})
}

// joins joins all input strings with a delimiter.
func joins(ec *EvalCtx, sep String, iterate func(func(Value))) {
	var buf bytes.Buffer
	iterate(func(v Value) {
		if s, ok := v.(String); ok {
			if buf.Len() > 0 {
				buf.WriteString(string(sep))
			}
			buf.WriteString(string(s))
		} else {
			throwf("join wants string input, got %s", v.Kind())
		}
	})
	out := ec.ports[1].Chan
	out <- String(buf.String())
}

// splits splits an argument strings by a delimiter and writes all pieces.
func splits(ec *EvalCtx, sep, s String) {
	out := ec.ports[1].Chan
	parts := strings.Split(string(s), string(sep))
	for _, p := range parts {
		out <- String(p)
	}
}

func hasPrefix(ec *EvalCtx, s, prefix String) {
	if !strings.HasPrefix(string(s), string(prefix)) {
		ec.falsify()
	}
}

func hasSuffix(ec *EvalCtx, s, suffix String) {
	if !strings.HasSuffix(string(s), string(suffix)) {
		ec.falsify()
	}
}

// toJSON converts a stream of Value's to JSON data.
func toJSON(ec *EvalCtx, iterate func(func(Value))) {
	out := ec.ports[1].File

	enc := json.NewEncoder(out)
	iterate(func(v Value) {
		err := enc.Encode(v)
		maybeThrow(err)
	})
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
func each(ec *EvalCtx, f FnValue, iterate func(func(Value))) {
	broken := false
	iterate(func(v Value) {
		if broken {
			return
		}
		// NOTE We don't have the position range of the closure in the source.
		// Ideally, it should be kept in the Closure itself.
		newec := ec.fork("closure of each")
		newec.ports[0] = NullClosedInput
		ex := newec.PCall(f, []Value{v}, NoOpts)
		ClosePorts(newec.ports)

		switch ex {
		case nil, Continue:
			// nop
		case Break:
			broken = true
		default:
			throw(ex)
		}
	})
}

// peach takes a single closure and applies it to all input values in parallel.
func peach(ec *EvalCtx, f FnValue, iterate func(func(Value))) {
	var w sync.WaitGroup

	broken := false
	var err error
	iterate(func(v Value) {
		if broken || err != nil {
			return
		}
		w.Add(1)
		go func() {
			// NOTE We don't have the position range of the closure in the source.
			// Ideally, it should be kept in the Closure itself.
			newec := ec.fork("closure of each")
			newec.ports[0] = NullClosedInput
			ex := newec.PCall(f, []Value{v}, NoOpts)
			ClosePorts(newec.ports)

			switch ex {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				err = ex
			}
			w.Done()
		}()
	})
	w.Wait()
	maybeThrow(err)
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
func eawk(ec *EvalCtx, f FnValue, iterate func(func(Value))) {
	broken := false
	iterate(func(v Value) {
		if broken {
			return
		}
		line, ok := v.(String)
		if !ok {
			throw(ErrInput)
		}
		args := []Value{line}
		for _, field := range eawkWordSep.Split(strings.Trim(string(line), " \t"), -1) {
			args = append(args, String(field))
		}

		newec := ec.fork("fn of eawk")
		// TODO: Close port 0 of newec.
		ex := newec.PCall(f, args, NoOpts)
		ClosePorts(newec.ports)

		switch ex {
		case nil, Continue:
			// nop
		case Break:
			broken = true
		default:
			throw(ex)
		}
	})
}

func constantly(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	// XXX Repr of this fn is not right
	out <- &BuiltinFn{
		"created by constantly",
		func(ec *EvalCtx, a []Value, o map[string]Value) {
			TakeNoOpt(o)
			if len(a) != 0 {
				throw(ErrArgs)
			}
			out := ec.ports[1].Chan
			for _, v := range args {
				out <- v
			}
		},
	}
}

func cd(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

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
	if ec.Store != nil {
		// XXX Error ignored.
		pwd, err := os.Getwd()
		if err == nil {
			store := ec.Store
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
	if ec.Store == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.Store.ListDirs()
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
	if ec.Store == nil {
		throw(ErrStoreNotConnected)
	}

	store := ec.Store
	seq, err := store.NextCmdSeq()
	maybeThrow(err)
	cmds, err := store.Cmds(0, seq)
	maybeThrow(err)

	out := ec.ports[1].Chan
	for _, cmd := range cmds {
		out <- String(cmd)
	}
}

func pathAbs(ec *EvalCtx, fname string) {
	out := ec.ports[1].Chan
	absname, err := filepath.Abs(fname)
	maybeThrow(err)
	out <- String(absname)
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
	if len(nums) == 0 {
		// Unary -
		sum = -sum
	} else {
		for _, f := range nums {
			sum -= f
		}
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

func slash(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	if len(args) == 0 {
		// cd /
		cdInner("/", ec)
		return
	}
	// Division
	wrappedDivide(ec, args, opts)
}

var wrappedDivide = WrapFn(divide)

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

func mod(ec *EvalCtx, a, b int) {
	out := ec.ports[1].Chan
	out <- String(strconv.Itoa(a % b))
}

func randFn(ec *EvalCtx) {
	out := ec.ports[1].Chan
	out <- String(fmt.Sprint(rand.Float64()))
}

func randint(ec *EvalCtx, low, high int) {
	if low >= high {
		throw(ErrArgs)
	}
	out := ec.ports[1].Chan
	i := low + rand.Intn(high-low)
	out <- String(strconv.Itoa(i))
}

func ord(ec *EvalCtx, s string) {
	out := ec.ports[1].Chan
	for _, r := range s {
		out <- String(fmt.Sprintf("0x%x", r))
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

func rangeFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var lower, upper int
	step := 1
	var err error

	switch len(args) {
	case 1:
		upper, err = toInt(args[0])
		maybeThrow(err)
	case 2, 3:
		lower, err = toInt(args[0])
		maybeThrow(err)
		upper, err = toInt(args[1])
		maybeThrow(err)
		if len(args) == 3 {
			step, err = toInt(args[2])
			maybeThrow(err)
		}
	default:
		throw(ErrArgs)
	}

	out := ec.ports[1].Chan
	for i := lower; i < upper; i += step {
		out <- String(strconv.Itoa(i))
	}
}

func boolFn(ec *EvalCtx, v Value) {
	out := ec.ports[1].Chan
	out <- Bool(ToBool(v))
}

func is(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	if len(args) < 2 {
		throw(ErrArgs)
	}
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			ec.falsify()
			return
		}
	}
}

func eq(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	if len(args) < 2 {
		throw(ErrArgs)
	}
	for i := 0; i+1 < len(args); i++ {
		if !DeepEq(args[i], args[i+1]) {
			ec.falsify()
			return
		}
	}
}

func resolveFn(ec *EvalCtx, cmd String) {
	out := ec.ports[1].Chan
	out <- resolve(string(cmd), ec)
}

func take(ec *EvalCtx, n int, iterate func(func(Value))) {
	out := ec.ports[1].Chan

	i := 0
	iterate(func(v Value) {
		if i < n {
			out <- v
		}
		i++
	})
}

func count(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var n int
	switch len(args) {
	case 0:
		// Count inputs.
		ec.IterateInputs(func(Value) {
			n++
		})
	case 1:
		// Get length of argument.
		v := args[0]
		if lener, ok := v.(Lener); ok {
			n = lener.Len()
		} else if iterator, ok := v.(Iterator); ok {
			iterator.Iterate(func(Value) bool {
				n++
				return true
			})
		} else {
			throw(fmt.Errorf("cannot get length of a %s", v.Kind()))
		}
	default:
		throw(errors.New("want 0 or 1 argument"))
	}
	ec.ports[1].Chan <- String(strconv.Itoa(n))
}

func wcswidth(ec *EvalCtx, s String) {
	out := ec.ports[1].Chan
	out <- String(strconv.Itoa(util.Wcswidth(string(s))))
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
			// TODO find command name
			errors[i] = Error{NewExternalCmdExit(fmt.Sprintf("(pid %d)", pid), ws, pid)}
		}
	}

	throwCompositeError(errors)
}

func tildeAbbr(ec *EvalCtx, path string) {
	out := ec.ports[1].Chan
	out <- String(util.TildeAbbr(path))
}

func fopen(ec *EvalCtx, name string) {
	// TODO support opening files for writing etc as well.
	out := ec.ports[1].Chan
	f, err := os.Open(name)
	maybeThrow(err)
	out <- File{f}
}

func pipe(ec *EvalCtx) {
	r, w, err := os.Pipe()
	out := ec.ports[1].Chan
	maybeThrow(err)
	out <- Pipe{r, w}
}

func fclose(ec *EvalCtx, f File)  { maybeThrow(f.inner.Close()) }
func prclose(ec *EvalCtx, p Pipe) { maybeThrow(p.r.Close()) }
func pwclose(ec *EvalCtx, p Pipe) { maybeThrow(p.w.Close()) }

func sleep(ec *EvalCtx, t float64) {
	d := time.Duration(float64(time.Second) * t)
	select {
	case <-ec.Interrupts():
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

func exec(ec *EvalCtx, args ...string) {
	if len(args) == 0 {
		args = []string{"elvish"}
	}
	var err error
	args[0], err = ec.Search(args[0])
	maybeThrow(err)

	preExit(ec)

	err = syscall.Exec(args[0], args, os.Environ())
	maybeThrow(err)
}

func exit(ec *EvalCtx, args ...int) {
	doexit := func(i int) {
		preExit(ec)
		os.Exit(i)
	}
	switch len(args) {
	case 0:
		doexit(0)
	case 1:
		doexit(args[0])
	default:
		throw(ErrArgs)
	}
}

func preExit(ec *EvalCtx) {
	err := ec.Store.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if ec.Stub != nil {
		ec.Stub.Terminate()
	}
}
