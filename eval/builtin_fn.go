package eval

// Builtin functions.

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"time"
	"unsafe"

	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hash"
)

// BuiltinFn is a builtin function.
type BuiltinFn struct {
	Name string
	Impl BuiltinFnImpl
}

type BuiltinFnImpl func(*EvalCtx, []Value, map[string]Value)

var _ CallableValue = &BuiltinFn{}

// Kind returns "fn".
func (*BuiltinFn) Kind() string {
	return "fn"
}

// Equal compares based on identity.
func (b *BuiltinFn) Equal(rhs interface{}) bool {
	return b == rhs
}

func (b *BuiltinFn) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(b))
}

// Repr returns an opaque representation "<builtin xxx>".
func (b *BuiltinFn) Repr(int) string {
	return "<builtin " + b.Name + ">"
}

// Call calls a builtin function.
func (b *BuiltinFn) Call(ec *EvalCtx, args []Value, opts map[string]Value) {
	b.Impl(ec, args, opts)
}

var builtinFns []*BuiltinFn

func addToBuiltinFns(moreFns []*BuiltinFn) {
	builtinFns = append(builtinFns, moreFns...)
}

// Builtins that have not been put into their own groups go here.

var ErrArgs = errors.New("args error")

func init() {
	addToBuiltinFns([]*BuiltinFn{
		{"nop", nop},

		{"kind-of", kindOf},

		{"is", is},
		{"eq", eq},
		{"not-eq", notEq},

		{"constantly", constantly},

		{"-source", source},

		{"bool", boolFn},
		{"not", not},

		// Time
		{"esleep", sleep},
		{"-time", _time},

		// Debugging
		{"-gc", _gc},
		{"-stack", _stack},
		{"-log", _log},
		// TODO(#327): Make this a variable, and make it possible to distinguish
		// filename ("/path/to/script.elv") and fabricated source name
		// ("[interactive]").
		{"-src-name", _getSrcName},

		{"-ifaddrs", _ifaddrs},
	})
	// For rand and randint.
	rand.Seed(time.Now().UTC().UnixNano())
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
	result := true
	for i := 0; i+1 < len(args); i++ {
		if !args[i].Equal(args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- Bool(result)
}

func notEq(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if args[i].Equal(args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- Bool(result)
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

	maybeThrow(ec.Source(string(fname)))
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

func _gc(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	runtime.GC()
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

func _getSrcName(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)
	ec.OutputChan() <- String(ec.srcName)
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
