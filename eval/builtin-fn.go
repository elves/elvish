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
	"net"
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
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var builtinFns []*BuiltinFn

// BuiltinFn is a builtin function.
type BuiltinFn struct {
	Name string
	Impl func(*EvalCtx, []Value, map[string]Value)
}

var _ CallableValue = &BuiltinFn{}

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
		// Trivial builtin
		&BuiltinFn{"nop", nop},

		// Introspection
		&BuiltinFn{"kind-of", kindOf},

		// Generic identity and equality
		&BuiltinFn{"is", is},
		&BuiltinFn{"eq", eq},

		// Value output
		&BuiltinFn{"put", put},

		// Bytes output
		&BuiltinFn{"print", print},
		&BuiltinFn{"echo", echo},
		&BuiltinFn{"pprint", pprint},

		// Bytes to value
		&BuiltinFn{"slurp", slurp},
		&BuiltinFn{"from-lines", fromLines},
		&BuiltinFn{"from-json", fromJSON},

		// Value to bytes
		&BuiltinFn{"to-lines", toLines},
		&BuiltinFn{"to-json", toJSON},

		// Exception and control
		&BuiltinFn{"fail", fail},
		&BuiltinFn{"multi-error", multiErrorFn},
		&BuiltinFn{"return", returnFn},
		&BuiltinFn{"break", breakFn},
		&BuiltinFn{"continue", continueFn},

		// Misc functional
		&BuiltinFn{"constantly", constantly},

		// Misc shell basic
		&BuiltinFn{"source", source},

		// Iterations.
		&BuiltinFn{"each", each},
		&BuiltinFn{"peach", peach},
		&BuiltinFn{"repeat", repeat},

		// Sequence primitives
		&BuiltinFn{"explode", explode},
		&BuiltinFn{"take", take},
		&BuiltinFn{"range", rangeFn},
		&BuiltinFn{"count", count},

		// String
		&BuiltinFn{"joins", joins},
		&BuiltinFn{"splits", splits},

		// String operations
		&BuiltinFn{"ord", ord},
		&BuiltinFn{"base", base},
		&BuiltinFn{"wcswidth", wcswidth},
		&BuiltinFn{"-override-wcwidth", overrideWcwidth},

		// String predicates
		&BuiltinFn{"has-prefix", hasPrefix},
		&BuiltinFn{"has-suffix", hasSuffix},

		// Regular expression
		&BuiltinFn{"-match", match},

		// String comparison
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

		// eawk
		&BuiltinFn{"eawk", eawk},

		// Directory
		&BuiltinFn{"cd", cd},
		&BuiltinFn{"dirs", dirs},

		// Path
		&BuiltinFn{"path-abs", wrapStringToStringError(filepath.Abs)},
		&BuiltinFn{"path-base", wrapStringToString(filepath.Base)},
		&BuiltinFn{"path-clean", wrapStringToString(filepath.Clean)},
		&BuiltinFn{"path-dir", wrapStringToString(filepath.Dir)},
		&BuiltinFn{"path-ext", wrapStringToString(filepath.Ext)},
		&BuiltinFn{"eval-symlinks", wrapStringToStringError(filepath.EvalSymlinks)},
		&BuiltinFn{"tilde-abbr", tildeAbbr},

		// Boolean operations
		&BuiltinFn{"bool", boolFn},
		&BuiltinFn{"not", not},
		&BuiltinFn{"true", trueFn},
		&BuiltinFn{"false", falseFn},

		// Arithmetics
		&BuiltinFn{"+", plus},
		&BuiltinFn{"-", minus},
		&BuiltinFn{"*", times},
		&BuiltinFn{"/", slash},
		&BuiltinFn{"^", pow},
		&BuiltinFn{"%", mod},

		// Random
		&BuiltinFn{"rand", randFn},
		&BuiltinFn{"randint", randint},

		// Numerical comparison
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

		// Command resolution
		&BuiltinFn{"resolve", resolveFn},
		&BuiltinFn{"has-external", hasExternal},
		&BuiltinFn{"search-external", searchExternal},

		// File and pipe
		&BuiltinFn{"fopen", fopen},
		&BuiltinFn{"fclose", fclose},
		&BuiltinFn{"pipe", pipe},
		&BuiltinFn{"prclose", prclose},
		&BuiltinFn{"pwclose", pwclose},

		// Process control
		&BuiltinFn{"fg", fg},
		&BuiltinFn{"exec", exec},
		&BuiltinFn{"exit", exit},

		// Time
		&BuiltinFn{"esleep", sleep},
		&BuiltinFn{"-time", _time},

		// Debugging
		&BuiltinFn{"-stack", _stack},
		&BuiltinFn{"-log", _log},

		&BuiltinFn{"-ifaddrs", _ifaddrs},
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
		result := true
		for i := 0; i < len(args)-1; i++ {
			if !cmp(string(args[i].(String)), string(args[i+1].(String))) {
				result = false
				break
			}
		}
		ec.OutputChan() <- Bool(result)
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
		result := true
		for i := 0; i < len(floats)-1; i++ {
			if !cmp(floats[i], floats[i+1]) {
				result = false
				break
			}
		}
		ec.OutputChan() <- Bool(result)
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

// ScanArgs scans arguments into pointers to supported argument types. If the
// arguments cannot be scanned, an error is thrown.
func ScanArgs(s []Value, args ...interface{}) {
	if len(s) != len(args) {
		throwf("arity mistmatch: want %d arguments, got %d", len(args), len(s))
	}
	for i, value := range s {
		scanArg(value, args[i])
	}
}

// ScanArgs is like ScanArgs, but the last element of args should be a pointer
// to a slice, and the rest of arguments will be scanned into it.
func ScanArgsVariadic(s []Value, args ...interface{}) {
	if len(s) < len(args)-1 {
		throwf("arity mistmatch: want at least %d arguments, got %d", len(args)-1, len(s))
	}
	ScanArgs(s[:len(args)-1], args[:len(args)-1]...)

	// Scan the rest of arguments into a slice.
	rest := s[len(args)-1:]
	dst := reflect.ValueOf(args[len(args)-1])
	if dst.Kind() != reflect.Ptr || dst.Elem().Kind() != reflect.Slice {
		throwf("internal bug: %T to ScanArgsVariadic, need pointer to slice", args[len(args)-1])
	}
	scanned := reflect.MakeSlice(dst.Elem().Type(), len(rest), len(rest))
	for i, value := range rest {
		scanArg(value, scanned.Index(i).Addr().Interface())
	}
	reflect.Indirect(dst).Set(scanned)
}

// ScanArgsAndOptionalIterate is like ScanArgs, but the argument can contain an
// optional iterable value at the end. The return value is a function that
// iterates the iterable value if it exists, or the input otherwise.
func ScanArgsAndOptionalIterate(ec *EvalCtx, s []Value, args ...interface{}) func(func(Value)) {
	switch len(s) {
	case len(args):
		ScanArgs(s, args...)
		return ec.IterateInputs
	case len(args) + 1:
		ScanArgs(s[:len(args)], args...)
		value := s[len(args)]
		iterable, ok := value.(Iterable)
		if !ok {
			throwf("need iterable argument, got %s", value.Kind())
		}
		return func(f func(Value)) {
			iterable.Iterate(func(v Value) bool {
				f(v)
				return true
			})
		}
	default:
		throwf("arity mistmatch: want %d or %d arguments, got %d", len(args), len(args)+1, len(s))
		return nil
	}
}

type Opt struct {
	Name    string
	Ptr     interface{}
	Default Value
}

func ScanOpts(m map[string]Value, opts ...Opt) {
	scanned := make(map[string]bool)
	for _, opt := range opts {
		a := opt.Ptr
		value, ok := m[opt.Name]
		if !ok {
			value = opt.Default
		}
		scanArg(value, a)
		scanned[opt.Name] = true
	}
	for key := range m {
		if !scanned[key] {
			throwf("unknown option %s", parse.Quote(key))
		}
	}
}

func scanArg(value Value, a interface{}) {
	ptr := reflect.ValueOf(a)
	if ptr.Kind() != reflect.Ptr {
		throwf("internal bug: %T to ScanArgs, need pointer", a)
	}
	v := reflect.Indirect(ptr)
	switch v.Kind() {
	case reflect.Int:
		i, err := toInt(value)
		maybeThrow(err)
		v.Set(reflect.ValueOf(i))
	case reflect.Float64:
		f, err := toFloat(value)
		maybeThrow(err)
		v.Set(reflect.ValueOf(f))
	default:
		if reflect.TypeOf(value).ConvertibleTo(v.Type()) {
			v.Set(reflect.ValueOf(value).Convert(v.Type()))
		} else {
			throwf("need %T argument, got %s", v.Interface(), value.Kind())
		}
	}
}

func nop(ec *EvalCtx, args []Value, opts map[string]Value) {
}

func kindOf(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- String(a.Kind())
	}
}

func is(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	if len(args) < 2 {
		throw(ErrArgs)
	}
	result := true
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			result = false
			break
		}
	}
	ec.OutputChan() <- Bool(result)
}

func eq(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	if len(args) < 2 {
		throw(ErrArgs)
	}
	result := true
	for i := 0; i+1 < len(args); i++ {
		if !DeepEq(args[i], args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- Bool(result)
}

func put(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- a
	}
}

func print(ec *EvalCtx, args []Value, opts map[string]Value) {
	var sepv String
	ScanOpts(opts, Opt{"sep", &sepv, String(" ")})

	out := ec.ports[1].File
	sep := string(sepv)
	for i, arg := range args {
		if i > 0 {
			out.WriteString(sep)
		}
		out.WriteString(ToString(arg))
	}
}

func echo(ec *EvalCtx, args []Value, opts map[string]Value) {
	print(ec, args, opts)
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

func slurp(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

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

func fromLines(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	in := ec.ports[0].File
	out := ec.ports[1].Chan

	linesToChan(in, out)
}

// fromJSON parses a stream of JSON data into Value's.
func fromJSON(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

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

func toLines(ec *EvalCtx, args []Value, opts map[string]Value) {
	iterate := ScanArgsAndOptionalIterate(ec, args)
	TakeNoOpt(opts)

	out := ec.ports[1].File

	iterate(func(v Value) {
		fmt.Fprintln(out, ToString(v))
	})
}

// toJSON converts a stream of Value's to JSON data.
func toJSON(ec *EvalCtx, args []Value, opts map[string]Value) {
	iterate := ScanArgsAndOptionalIterate(ec, args)
	TakeNoOpt(opts)

	out := ec.ports[1].File

	enc := json.NewEncoder(out)
	iterate(func(v Value) {
		err := enc.Encode(v)
		maybeThrow(err)
	})
}

func fail(ec *EvalCtx, args []Value, opts map[string]Value) {
	var msg String
	ScanArgs(args, &msg)
	TakeNoOpt(opts)

	throw(errors.New(string(msg)))
}

func multiErrorFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	var excs []*Exception
	ScanArgsVariadic(args, &excs)
	TakeNoOpt(opts)

	throw(PipelineError{excs})
}

func returnFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	throw(Return)
}

func breakFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	throw(Break)
}

func continueFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	throw(Continue)
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

func source(ec *EvalCtx, args []Value, opts map[string]Value) {
	var fname String
	ScanArgs(args, &fname)
	ScanOpts(opts)

	ec.Source(string(fname))
}

// each takes a single closure and applies it to all input values.
func each(ec *EvalCtx, args []Value, opts map[string]Value) {
	var f CallableValue
	iterate := ScanArgsAndOptionalIterate(ec, args, &f)
	TakeNoOpt(opts)

	broken := false
	iterate(func(v Value) {
		if broken {
			return
		}
		// NOTE We don't have the position range of the closure in the source.
		// Ideally, it should be kept in the Closure itself.
		newec := ec.fork("closure of each")
		newec.ports[0] = DevNullClosedChan
		ex := newec.PCall(f, []Value{v}, NoOpts)
		ClosePorts(newec.ports)

		if ex != nil {
			switch ex.(*Exception).Cause {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				throw(ex)
			}
		}
	})
}

// peach takes a single closure and applies it to all input values in parallel.
func peach(ec *EvalCtx, args []Value, opts map[string]Value) {
	var f CallableValue
	iterate := ScanArgsAndOptionalIterate(ec, args, &f)
	TakeNoOpt(opts)

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
			newec.ports[0] = DevNullClosedChan
			ex := newec.PCall(f, []Value{v}, NoOpts)
			ClosePorts(newec.ports)

			if ex != nil {
				switch ex.(*Exception).Cause {
				case nil, Continue:
					// nop
				case Break:
					broken = true
				default:
					err = ex
				}
			}
			w.Done()
		}()
	})
	w.Wait()
	maybeThrow(err)
}

func repeat(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		n int
		v Value
	)
	ScanArgs(args, &n, &v)
	TakeNoOpt(opts)

	out := ec.OutputChan()
	for i := 0; i < n; i++ {
		out <- v
	}
}

// explode puts each element of the argument.
func explode(ec *EvalCtx, args []Value, opts map[string]Value) {
	var v IterableValue
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	v.Iterate(func(e Value) bool {
		out <- e
		return true
	})
}

func take(ec *EvalCtx, args []Value, opts map[string]Value) {
	var n int
	iterate := ScanArgsAndOptionalIterate(ec, args, &n)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	i := 0
	iterate(func(v Value) {
		if i < n {
			out <- v
		}
		i++
	})
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
		} else if iterator, ok := v.(Iterable); ok {
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

// joins joins all input strings with a delimiter.
func joins(ec *EvalCtx, args []Value, opts map[string]Value) {
	var sepv String
	iterate := ScanArgsAndOptionalIterate(ec, args, &sepv)
	sep := string(sepv)
	TakeNoOpt(opts)

	var buf bytes.Buffer
	iterate(func(v Value) {
		if s, ok := v.(String); ok {
			if buf.Len() > 0 {
				buf.WriteString(sep)
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
func splits(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s, sep String
	ScanArgs(args, &s)
	ScanOpts(opts, Opt{"sep", &sep, String("")})

	out := ec.ports[1].Chan
	parts := strings.Split(string(s), string(sep))
	for _, p := range parts {
		out <- String(p)
	}
}

func ord(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s String
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	for _, r := range s {
		out <- String(fmt.Sprintf("0x%x", r))
	}
}

var ErrBadBase = errors.New("bad base")

func base(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		b    int
		nums []int
	)
	ScanArgsVariadic(args, &b, &nums)
	TakeNoOpt(opts)

	if b < 2 || b > 36 {
		throw(ErrBadBase)
	}

	out := ec.ports[1].Chan

	for _, num := range nums {
		out <- String(strconv.FormatInt(int64(num), b))
	}
}

func wcswidth(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s String
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(strconv.Itoa(util.Wcswidth(string(s))))
}

func overrideWcwidth(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		s String
		w int
	)
	ScanArgs(args, &s, &w)
	TakeNoOpt(opts)

	r, err := toRune(s)
	maybeThrow(err)
	util.OverrideWcwidth(r, w)
}

func hasPrefix(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s, prefix String
	ScanArgs(args, &s, &prefix)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(strings.HasPrefix(string(s), string(prefix)))
}

func hasSuffix(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s, suffix String
	ScanArgs(args, &s, &suffix)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(strings.HasSuffix(string(s), string(suffix)))
}

func match(ec *EvalCtx, args []Value, opts map[string]Value) {
	var p, s String
	ScanArgs(args, &p, &s)
	TakeNoOpt(opts)

	matched, err := regexp.MatchString(string(p), string(s))
	maybeThrow(err)
	ec.OutputChan() <- Bool(matched)
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
func eawk(ec *EvalCtx, args []Value, opts map[string]Value) {
	var f CallableValue
	iterate := ScanArgsAndOptionalIterate(ec, args, &f)
	TakeNoOpt(opts)

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

		if ex != nil {
			switch ex.(*Exception).Cause {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				throw(ex)
			}
		}
	})
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

func dirs(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	if ec.Store == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.Store.GetDirs(store.NoBlacklist)
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

func tildeAbbr(ec *EvalCtx, args []Value, opts map[string]Value) {
	var pathv String
	ScanArgs(args, &pathv)
	path := string(pathv)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(util.TildeAbbr(path))
}

func boolFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	var v Value
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(ToBool(v))
}

func not(ec *EvalCtx, args []Value, opts map[string]Value) {
	var v Value
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(!ToBool(v))
}

func trueFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	ec.OutputChan() <- Bool(true)
}

func falseFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	ec.OutputChan() <- Bool(false)
}

func plus(ec *EvalCtx, args []Value, opts map[string]Value) {
	var nums []float64
	ScanArgsVariadic(args, &nums)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	out <- String(fmt.Sprintf("%g", sum))
}

func minus(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		sum  float64
		nums []float64
	)
	ScanArgsVariadic(args, &sum, &nums)
	TakeNoOpt(opts)

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

func times(ec *EvalCtx, args []Value, opts map[string]Value) {
	var nums []float64
	ScanArgsVariadic(args, &nums)
	TakeNoOpt(opts)

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
	divide(ec, args, opts)
}

func divide(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		prod float64
		nums []float64
	)
	ScanArgsVariadic(args, &prod, &nums)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	for _, f := range nums {
		prod /= f
	}
	out <- String(fmt.Sprintf("%g", prod))
}

func pow(ec *EvalCtx, args []Value, opts map[string]Value) {
	var b, p float64
	ScanArgs(args, &b, &p)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(fmt.Sprintf("%g", math.Pow(b, p)))
}

func mod(ec *EvalCtx, args []Value, opts map[string]Value) {
	var a, b int
	ScanArgs(args, &a, &b)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(strconv.Itoa(a % b))
}

func randFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(fmt.Sprint(rand.Float64()))
}

func randint(ec *EvalCtx, args []Value, opts map[string]Value) {
	var low, high int
	ScanArgs(args, &low, &high)
	TakeNoOpt(opts)

	if low >= high {
		throw(ErrArgs)
	}
	out := ec.ports[1].Chan
	i := low + rand.Intn(high-low)
	out <- String(strconv.Itoa(i))
}

func resolveFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	var cmd String
	ScanArgs(args, &cmd)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- resolve(string(cmd), ec)
}

func hasExternal(ec *EvalCtx, args []Value, opts map[string]Value) {
	var cmd String
	ScanArgs(args, &cmd)
	TakeNoOpt(opts)

	_, err := ec.Search(string(cmd))
	ec.OutputChan() <- Bool(err == nil)
}

func searchExternal(ec *EvalCtx, args []Value, opts map[string]Value) {
	var cmd String
	ScanArgs(args, &cmd)
	TakeNoOpt(opts)

	path, err := ec.Search(string(cmd))
	maybeThrow(err)

	out := ec.ports[1].Chan
	out <- String(path)
}

func fopen(ec *EvalCtx, args []Value, opts map[string]Value) {
	var namev String
	ScanArgs(args, &namev)
	name := string(namev)
	TakeNoOpt(opts)

	// TODO support opening files for writing etc as well.
	out := ec.ports[1].Chan
	f, err := os.Open(name)
	maybeThrow(err)
	out <- File{f}
}

func fclose(ec *EvalCtx, args []Value, opts map[string]Value) {
	var f File
	ScanArgs(args, &f)
	TakeNoOpt(opts)

	maybeThrow(f.inner.Close())
}

func pipe(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	r, w, err := os.Pipe()
	out := ec.ports[1].Chan
	maybeThrow(err)
	out <- Pipe{r, w}
}

func prclose(ec *EvalCtx, args []Value, opts map[string]Value) {
	var p Pipe
	ScanArgs(args, &p)
	TakeNoOpt(opts)

	maybeThrow(p.r.Close())
}

func pwclose(ec *EvalCtx, args []Value, opts map[string]Value) {
	var p Pipe
	ScanArgs(args, &p)
	TakeNoOpt(opts)

	maybeThrow(p.w.Close())
}

func fg(ec *EvalCtx, args []Value, opts map[string]Value) {
	var pids []int
	ScanArgsVariadic(args, &pids)
	TakeNoOpt(opts)

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

	errors := make([]*Exception, len(pids))

	for i, pid := range pids {
		err := syscall.Kill(pid, syscall.SIGCONT)
		if err != nil {
			errors[i] = &Exception{err, nil}
		}
	}

	for i, pid := range pids {
		if errors[i] != nil {
			continue
		}
		var ws syscall.WaitStatus
		_, err = syscall.Wait4(pid, &ws, syscall.WUNTRACED, nil)
		if err != nil {
			errors[i] = &Exception{err, nil}
		} else {
			// TODO find command name
			errors[i] = &Exception{NewExternalCmdExit(fmt.Sprintf("(pid %d)", pid), ws, pid), nil}
		}
	}

	maybeThrow(ComposeExceptionsFromPipeline(errors))
}

func exec(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var argstrings []string
	if len(args) == 0 {
		argstrings = []string{"elvish"}
	} else {
		argstrings = make([]string, len(args))
		for i, a := range args {
			argstrings[i] = ToString(a)
		}
	}

	var err error
	argstrings[0], err = ec.Search(argstrings[0])
	maybeThrow(err)

	preExit(ec)

	err = syscall.Exec(argstrings[0], argstrings, os.Environ())
	maybeThrow(err)
}

func exit(ec *EvalCtx, args []Value, opts map[string]Value) {
	var codes []int
	ScanArgsVariadic(args, &codes)
	TakeNoOpt(opts)

	doexit := func(i int) {
		preExit(ec)
		os.Exit(i)
	}
	switch len(codes) {
	case 0:
		doexit(0)
	case 1:
		doexit(codes[0])
	default:
		throw(ErrArgs)
	}
}

func sleep(ec *EvalCtx, args []Value, opts map[string]Value) {
	var t float64
	ScanArgs(args, &t)
	TakeNoOpt(opts)

	d := time.Duration(float64(time.Second) * t)
	select {
	case <-ec.Interrupts():
		throw(ErrInterrupted)
	case <-time.After(d):
	}
}

func _time(ec *EvalCtx, args []Value, opts map[string]Value) {
	var f CallableValue
	ScanArgs(args, &f)
	TakeNoOpt(opts)

	t0 := time.Now()
	f.Call(ec, NoArgs, NoOpts)
	t1 := time.Now()

	dt := t1.Sub(t0)
	fmt.Fprintln(ec.ports[1].File, dt)
}

func _stack(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	out := ec.ports[1].File
	// XXX dup with main.go
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)
}

func _log(ec *EvalCtx, args []Value, opts map[string]Value) {
	var fnamev String
	ScanArgs(args, &fnamev)
	fname := string(fnamev)
	TakeNoOpt(opts)

	maybeThrow(util.SetOutputFile(fname))
}

func _ifaddrs(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan

	addrs, err := net.InterfaceAddrs()
	maybeThrow(err)
	for _, addr := range addrs {
		out <- String(addr.String())
	}
}

func toFloat(arg Value) (float64, error) {
	if _, ok := arg.(String); !ok {
		return 0, fmt.Errorf("must be string")
	}
	s := string(arg.(String))
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		num, err2 := strconv.ParseInt(s, 0, 64)
		if err2 != nil {
			return 0, err
		}
		return float64(num), nil
	}
	return num, nil
}

func toInt(arg Value) (int, error) {
	arg, ok := arg.(String)
	if !ok {
		return 0, fmt.Errorf("must be string")
	}
	num, err := strconv.ParseInt(string(arg.(String)), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func toRune(arg Value) (rune, error) {
	ss, ok := arg.(String)
	if !ok {
		return -1, fmt.Errorf("must be string")
	}
	s := string(ss)
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return -1, fmt.Errorf("string is not valid UTF-8")
	}
	if size != len(s) {
		return -1, fmt.Errorf("string has multiple runes")
	}
	return r, nil
}

func preExit(ec *EvalCtx) {
	err := ec.Store.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
