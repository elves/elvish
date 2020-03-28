package eval

import (
	"fmt"
	"strconv"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/util"
)

// Unwrappers are helper types for "unwrapping" values, the process for
// asserting certain properties of values and throwing exceptions when such
// properties are not satisfied.

type unwrapper struct {
	// ctx is the evaluation context.
	ctx *Frame
	// description describes what is being unwrapped. It is used in error
	// messages.
	description string
	// begin and end contains positions in the source code to point to when
	// error occurs.
	begin, end int
	// values contain the Value's to unwrap.
	values []interface{}
	// Any errors during the unwrapping.
	err error
}

func (u *unwrapper) error(want, gotfmt string, gotargs ...interface{}) {
	if u.err != nil {
		return
	}
	got := fmt.Sprintf(gotfmt, gotargs...)
	u.err = u.ctx.errorpf(
		u.begin, u.end, "%s must be %s; got %s", u.description, want, got)
}

// ValuesUnwrapper unwraps []Value.
type ValuesUnwrapper struct{ *unwrapper }

// ExecAndUnwrap executes a ValuesOp and creates an Unwrapper for the obtained
// values.
func (ctx *Frame) ExecAndUnwrap(desc string, op valuesOp) ValuesUnwrapper {
	values, err := op.exec(ctx)
	return ValuesUnwrapper{&unwrapper{ctx, desc, op.From, op.To, values, err}}
}

// One unwraps the value to be exactly one value.
func (u ValuesUnwrapper) One() ValueUnwrapper {
	if len(u.values) != 1 {
		u.error("a single value", "%d values", len(u.values))
	}
	return ValueUnwrapper{u.unwrapper}
}

// ValueUnwrapper unwraps one Value.
type ValueUnwrapper struct{ *unwrapper }

func (u ValueUnwrapper) Any() (interface{}, error) {
	if u.err != nil {
		return nil, u.err
	}
	return u.values[0], u.err
}

func (u ValueUnwrapper) String() (string, error) {
	if u.err != nil {
		return "", u.err
	}
	s, ok := u.values[0].(string)
	if !ok {
		u.error("string", "%s", vals.Kind(u.values[0]))
	}
	return s, u.err
}

func (u ValueUnwrapper) Int() (int, error) {
	s, err := u.String()
	if err != nil {
		return 0, u.err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		u.error("integer", "%s", s)
	}
	return i, u.err
}

func (u ValueUnwrapper) NonNegativeInt() (int, error) {
	i, err := u.Int()
	if err != nil {
		return 0, err
	}
	if i < 0 {
		u.error("non-negative int", "%d", i)
	}
	return i, u.err
}

func (u ValueUnwrapper) Fd() (int, error) {
	s, err := u.String()
	if err != nil {
		return 0, err
	}
	switch s {
	case "stdin":
		return 0, nil
	case "stdout":
		return 1, nil
	case "stderr":
		return 2, nil
	default:
		i, err := u.NonNegativeInt()
		if err != nil {
			return 0, u.ctx.errorpf(u.begin, u.end, "fd must be standard stream name or integer; got %s", s)
		}
		return i, nil
	}
}

func (u ValueUnwrapper) FdOrClose() (int, error) {
	s, err := u.String()
	if err == nil && s == "-" {
		return -1, nil
	}
	fd, err := u.Fd()
	if err != nil {
		return 0, u.ctx.errorpf(u.begin, u.end, "redirection source must be standard stream name or integer; got %s", s)
	}
	return fd, nil
}

func (u ValueUnwrapper) CommandHead() (Callable, error) {
	if u.err != nil {
		return nil, u.err
	}
	switch v := u.values[0].(type) {
	case Callable:
		return v, nil
	case string:
		if util.DontSearch(v) {
			return ExternalCmd{v}, nil
		}
	}
	u.error("callable or relative path", "%s", vals.Kind(u.values[0]))
	return nil, u.err
}
