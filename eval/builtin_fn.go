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

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hash"
)

// BuiltinFn is a builtin function.
type BuiltinFn struct {
	Name string
	Impl BuiltinFnImpl
}

type BuiltinFnImpl func(*Frame, []types.Value, map[string]types.Value)

var _ Fn = &BuiltinFn{}

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
func (b *BuiltinFn) Call(ec *Frame, args []types.Value, opts map[string]types.Value) {
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

		{"bool", boolFn},
		{"not", not},
		{"is", is},
		{"eq", eq},
		{"not-eq", notEq},

		{"constantly", constantly},

		{"-source", source},

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

func nop(ec *Frame, args []types.Value, opts map[string]types.Value) {
}

func kindOf(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- String(a.Kind())
	}
}

func boolFn(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var v types.Value
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(types.ToBool(v))
}

func not(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var v types.Value
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(!types.ToBool(v))
}

func is(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func eq(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if !args[i].Equal(args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func notEq(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if args[i].Equal(args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func constantly(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	// XXX Repr of this fn is not right
	out <- &BuiltinFn{
		"created by constantly",
		func(ec *Frame, a []types.Value, o map[string]types.Value) {
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

func source(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var fname String
	ScanArgs(args, &fname)
	ScanOpts(opts)

	maybeThrow(ec.Source(string(fname)))
}

func sleep(ec *Frame, args []types.Value, opts map[string]types.Value) {
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

func _time(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var f Fn
	ScanArgs(args, &f)
	TakeNoOpt(opts)

	t0 := time.Now()
	f.Call(ec, NoArgs, NoOpts)
	t1 := time.Now()

	dt := t1.Sub(t0)
	fmt.Fprintln(ec.ports[1].File, dt)
}

func _gc(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	runtime.GC()
}

func _stack(ec *Frame, args []types.Value, opts map[string]types.Value) {
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

func _log(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var fnamev String
	ScanArgs(args, &fnamev)
	fname := string(fnamev)
	TakeNoOpt(opts)

	maybeThrow(util.SetOutputFile(fname))
}

func _getSrcName(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)
	ec.OutputChan() <- String(ec.srcName)
}

func _ifaddrs(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan

	addrs, err := net.InterfaceAddrs()
	maybeThrow(err)
	for _, addr := range addrs {
		out <- String(addr.String())
	}
}
