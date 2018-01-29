package eval

// Builtin functions.

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
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

type BuiltinFnImpl func(*Frame, []interface{}, map[string]interface{})

var _ Callable = &BuiltinFn{}

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
func (b *BuiltinFn) Call(ec *Frame, args []interface{}, opts map[string]interface{}) error {
	return util.PCall(func() { b.Impl(ec, args, opts) })
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
		{"src", src},
		{"-gc", _gc},
		{"-stack", _stack},
		{"-log", _log},

		{"-ifaddrs", _ifaddrs},
	})
	// For rand and randint.
	rand.Seed(time.Now().UTC().UnixNano())
}

func nop(ec *Frame, args []interface{}, opts map[string]interface{}) {
}

func kindOf(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	out := ec.ports[1].Chan
	for _, a := range args {
		out <- types.Kind(a)
	}
}

func boolFn(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var v interface{}
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(v)
}

func not(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var v interface{}
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- !types.Bool(v)
}

func is(ec *Frame, args []interface{}, opts map[string]interface{}) {
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

func eq(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if !types.Equal(args[i], args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func notEq(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if types.Equal(args[i], args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func constantly(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	// XXX Repr of this fn is not right
	out <- &BuiltinFn{
		"created by constantly",
		func(ec *Frame, a []interface{}, o map[string]interface{}) {
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

func source(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var argFname string
	ScanArgs(args, &argFname)
	ScanOpts(opts)

	fname := argFname
	abs, err := filepath.Abs(fname)
	maybeThrow(err)

	maybeThrow(ec.Source(fname, abs))
}

func sleep(ec *Frame, args []interface{}, opts map[string]interface{}) {
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

func _time(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var f Callable
	ScanArgs(args, &f)
	TakeNoOpt(opts)

	t0 := time.Now()
	err := f.Call(ec, NoArgs, NoOpts)
	maybeThrow(err)
	t1 := time.Now()

	dt := t1.Sub(t0)
	fmt.Fprintln(ec.ports[1].File, dt)
}

func src(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	ec.OutputChan() <- ec.srcMeta
}

func _gc(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	runtime.GC()
}

func _stack(ec *Frame, args []interface{}, opts map[string]interface{}) {
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

func _log(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var fnamev string
	ScanArgs(args, &fnamev)
	fname := fnamev
	TakeNoOpt(opts)

	maybeThrow(util.SetOutputFile(fname))
}

func _ifaddrs(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan

	addrs, err := net.InterfaceAddrs()
	maybeThrow(err)
	for _, addr := range addrs {
		out <- addr.String()
	}
}
