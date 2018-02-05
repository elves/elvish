package eval

// Builtin functions.

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
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
	addToReflectBuiltinFns(map[string]interface{}{
		"nop":        nop,
		"kind-of":    kindOf,
		"constantly": constantly,

		"-source": source,

		// Time
		"esleep": sleep,
		"-time":  _time,

		"-ifaddrs": _ifaddrs,
	})

	// For rand and randint.
	rand.Seed(time.Now().UTC().UnixNano())
}

func nop(opts Options, args ...interface{}) {
	// Do nothing
}

func kindOf(fm *Frame, args ...interface{}) {
	out := fm.ports[1].Chan
	for _, a := range args {
		out <- types.Kind(a)
	}
}

func constantly(args ...interface{}) Callable {
	// XXX Repr of this fn is not right
	return NewReflectBuiltinFn(
		"created by constantly",
		func(fm *Frame) {
			out := fm.ports[1].Chan
			for _, v := range args {
				out <- v
			}
		},
	)
}

func source(fm *Frame, fname string) error {
	abs, err := filepath.Abs(fname)
	if err != nil {
		return err
	}
	return fm.Source(fname, abs)
}

func sleep(fm *Frame, t float64) error {
	d := time.Duration(float64(time.Second) * t)
	select {
	case <-fm.Interrupts():
		return ErrInterrupted
	case <-time.After(d):
		return nil
	}
}

func _time(ec *Frame, f Callable) error {
	t0 := time.Now()
	err := f.Call(ec, NoArgs, NoOpts)
	t1 := time.Now()

	dt := t1.Sub(t0)
	fmt.Fprintln(ec.ports[1].File, dt)

	return err
}

func _ifaddrs(fm *Frame) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	out := fm.ports[1].Chan
	for _, addr := range addrs {
		out <- addr.String()
	}
	return nil
}
