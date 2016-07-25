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
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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
	return "<builtin " + b.Name + ">"
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
		&BuiltinFn{"false", falseFn},

		&BuiltinFn{"print", WrapFn(print)},
		&BuiltinFn{"println", WrapFn(println)},
		&BuiltinFn{"pprint", pprint},

		&BuiltinFn{"slurp", WrapFn(slurp)},
		&BuiltinFn{"into-lines", WrapFn(intoLines)},

		&BuiltinFn{"rat", WrapFn(ratFn)},

		&BuiltinFn{"put", put},
		&BuiltinFn{"unpack", WrapFn(unpack)},

		&BuiltinFn{"joins", WrapFn(joins)},
		&BuiltinFn{"splits", WrapFn(splits)},
		&BuiltinFn{"has-prefix", WrapFn(hasPrefix)},
		&BuiltinFn{"has-suffix", WrapFn(hasSuffix)},

		&BuiltinFn{"to-json", WrapFn(toJSON)},
		&BuiltinFn{"from-json", WrapFn(fromJSON)},

		&BuiltinFn{"kind-of", kindOf},

		&BuiltinFn{"fail", WrapFn(fail)},
		&BuiltinFn{"multi-error", WrapFn(multiErrorFn)},
		&BuiltinFn{"return", WrapFn(returnFn)},
		&BuiltinFn{"break", WrapFn(breakFn)},
		&BuiltinFn{"continue", WrapFn(continueFn)},

		&BuiltinFn{"each", WrapFn(each)},
		&BuiltinFn{"eawk", WrapFn(eawk)},

		&BuiltinFn{"cd", cd},
		&BuiltinFn{"dirs", WrapFn(dirs)},
		&BuiltinFn{"history", WrapFn(history)},

		&BuiltinFn{"source", WrapFn(source)},

		&BuiltinFn{"+", WrapFn(plus)},
		&BuiltinFn{"-", WrapFn(minus)},
		&BuiltinFn{"*", WrapFn(times)},
		&BuiltinFn{"/", slash},
		&BuiltinFn{"^", WrapFn(pow)},
		&BuiltinFn{"<", WrapFn(lt)},
		&BuiltinFn{"<=", WrapFn(le)},
		&BuiltinFn{">", WrapFn(gt)},
		&BuiltinFn{">=", WrapFn(ge)},
		&BuiltinFn{"%", WrapFn(mod)},
		&BuiltinFn{"rand", WrapFn(randFn)},
		&BuiltinFn{"randint", WrapFn(randint)},

		&BuiltinFn{"ord", WrapFn(ord)},
		&BuiltinFn{"base", WrapFn(base)},

		&BuiltinFn{"bool", WrapFn(boolFn)},
		&BuiltinFn{"==", eq},
		&BuiltinFn{"!=", WrapFn(noteq)},
		&BuiltinFn{"deepeq", deepeq},

		&BuiltinFn{"resolve", WrapFn(resolveFn)},

		&BuiltinFn{"take", WrapFn(take)},

		&BuiltinFn{"count", count},
		&BuiltinFn{"rest", WrapFn(rest)},

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
// generates argument checking and conversion code according to the signature
// of the inner function. The inner function must accept evalCtx* as the first
// argument and return an exitus.
func WrapFn(inner interface{}) func(*EvalCtx, []Value) {
	funcType := reflect.TypeOf(inner)
	if funcType.In(0) != evalCtxType {
		panic("bad func to wrap, first argument not *EvalCtx")
	}

	fixedArgs := funcType.NumIn() - 1
	isVariadic := funcType.IsVariadic()
	hasOptionalIterate := false
	var variadicType reflect.Type
	if isVariadic {
		fixedArgs--
		variadicType = funcType.In(funcType.NumIn() - 1).Elem()
		if !supportedArgType(variadicType) {
			panic(fmt.Sprintf("bad func to wrap, variadic argument type %s unsupported", variadicType))
		}
	} else if funcType.In(funcType.NumIn()-1) == iterateType {
		fixedArgs--
		hasOptionalIterate = true
	}

	for i := 0; i < fixedArgs; i++ {
		if !supportedArgType(funcType.In(i + 1)) {
			panic(fmt.Sprintf("bad func to wrap, argument type %s unsupported", funcType.In(i+1)))
		}
	}

	return func(ec *EvalCtx, args []Value) {
		if isVariadic {
			if len(args) < fixedArgs {
				throw(fmt.Errorf("arity mismatch: want at least %d arguments, got %d", fixedArgs, len(args)))
			}
		} else if hasOptionalIterate {
			if len(args) < fixedArgs || len(args) > fixedArgs+1 {
				throw(fmt.Errorf("arity mismatch: want %d or %d arguments, got %d", fixedArgs, fixedArgs+1, len(args)))
			}
		} else if len(args) != fixedArgs {
			throw(fmt.Errorf("arity mismatch: want %d arguments, got %d", fixedArgs, len(args)))
		}
		convertedArgs := make([]reflect.Value, len(args)+1)
		convertedArgs[0] = reflect.ValueOf(ec)

		var err error
		for i, arg := range args[:fixedArgs] {
			convertedArgs[1+i], err = convertArg(arg, funcType.In(i+1))
			if err != nil {
				throw(errors.New("bad argument: " + err.Error()))
			}
		}

		if isVariadic {
			for i, arg := range args[fixedArgs:] {
				convertedArgs[1+fixedArgs+i], err = convertArg(arg, variadicType)
				if err != nil {
					throw(errors.New("bad argument: " + err.Error()))
				}
			}
		} else if hasOptionalIterate {
			var iterate func(func(Value))
			if len(args) == fixedArgs {
				// No Elemser specified in arguments. Use input.
				// Since convertedArgs was created according to the size of the
				// actual argument list, we now an empty element to make room
				// for this additional iterator argument.
				convertedArgs = append(convertedArgs, reflect.Value{})
				iterate = ec.IterateInputs
			} else {
				iterator, ok := args[fixedArgs].(Iterator)
				if !ok {
					throw(errors.New("bad argument: need iterator, got " + args[fixedArgs].Kind()))
				}
				iterate = func(f func(Value)) {
					iterator.Iterate(func(v Value) bool {
						f(v)
						return true
					})
				}
			}
			convertedArgs[1+fixedArgs] = reflect.ValueOf(iterate)
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

func nop(ec *EvalCtx, args []Value) {
}

func falseFn(ec *EvalCtx, args []Value) {
	ec.falsify()
}

func put(ec *EvalCtx, args []Value) {
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- a
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

func pprint(ec *EvalCtx, args []Value) {
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
	maybeThrow(err)
	out <- String(string(all))
}

func intoLines(ec *EvalCtx, iterate func(func(Value))) {
	out := ec.ports[1].File

	iterate(func(v Value) {
		fmt.Fprintln(out, ToString(v))
	})
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
func unpack(ec *EvalCtx, iterate func(func(Value))) {
	out := ec.ports[1].Chan

	iterate(func(v Value) {
		iterator, ok := v.(Iterator)
		if !ok {
			throwf("unpack wants iterator in input, got %s", v.Kind())
		}
		iterator.Iterate(func(v Value) bool {
			out <- v
			return true
		})
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
// TODO: make sep an option. (`split $s &sep=:` instead of `split $s :`)
func splits(ec *EvalCtx, s, sep String) {
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
		// TODO: Close port 0 of newec.
		ex := newec.PCall(f, []Value{v})
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
		ex := newec.PCall(f, args)
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

func slash(ec *EvalCtx, args []Value) {
	if len(args) == 0 {
		// cd /
		cdInner("/", ec)
		return
	}
	// Division
	wrappedDivide(ec, args)
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

func lt(ec *EvalCtx, nums ...float64) {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] < nums[i+1]) {
			ec.falsify()
		}
	}
}

func le(ec *EvalCtx, nums ...float64) {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] <= nums[i+1]) {
			ec.falsify()
		}
	}
}

func gt(ec *EvalCtx, nums ...float64) {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] > nums[i+1]) {
			ec.falsify()
		}
	}
}

func ge(ec *EvalCtx, nums ...float64) {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] >= nums[i+1]) {
			ec.falsify()
		}
	}
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

func boolFn(ec *EvalCtx, v Value) {
	out := ec.ports[1].Chan
	out <- Bool(ToBool(v))
}

func eq(ec *EvalCtx, args []Value) {
	if len(args) == 0 {
		ec.falsify()
		return
	}
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			ec.falsify()
			return
		}
	}
}

func noteq(ec *EvalCtx, lhs, rhs Value) {
	if lhs == rhs {
		ec.falsify()
	}
}

func deepeq(ec *EvalCtx, args []Value) {
	if len(args) == 0 {
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

func count(ec *EvalCtx, args []Value) {
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

func rest(ec *EvalCtx, li List) {
	out := ec.ports[1].Chan
	restli := (*li.inner)[1:]
	out <- List{&restli}
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
	err := ec.store.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if ec.Stub != nil {
		ec.Stub.Terminate()
	}
}
