package eval

// Builtin functions.

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
	"time"

	"github.com/elves/elvish/eval/vals"
)

// Builtins that have not been put into their own groups go here.

var ErrArgs = errors.New("args error")

func init() {
	addBuiltinFns(map[string]interface{}{
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

func nop(opts RawOptions, args ...interface{}) {
	// Do nothing
}

func kindOf(fm *Frame, args ...interface{}) {
	out := fm.ports[1].Chan
	for _, a := range args {
		out <- vals.Kind(a)
	}
}

func constantly(args ...interface{}) Callable {
	// XXX Repr of this fn is not right
	return NewBuiltinFn(
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
