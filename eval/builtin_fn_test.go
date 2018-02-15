package eval

import (
	"errors"
	"testing"

	"github.com/elves/elvish/eval/vals"
)

func TestReflectBuiltinFnCall(t *testing.T) {
	theFrame := new(Frame)
	theOptions := map[string]interface{}{}

	var f Callable
	callGood := func(fm *Frame, args []interface{}, opts map[string]interface{}) {
		err := f.Call(fm, args, opts)
		if err != nil {
			t.Errorf("Failed to call f: %v", err)
		}
	}
	callBad := func(fm *Frame, args []interface{}, opts map[string]interface{}) {
		err := f.Call(fm, args, opts)
		if err == nil {
			t.Errorf("Calling f didn't return error")
		}
	}

	// *Frame parameter gets the Frame.
	f = NewBuiltinFn("f", func(f *Frame) {
		if f != theFrame {
			t.Errorf("*Frame parameter doesn't get current frame")
		}
	})
	callGood(theFrame, nil, theOptions)

	// Options parameter gets options.
	f = NewBuiltinFn("f", func(opts Options) {
		if opts["foo"] != "bar" {
			t.Errorf("Options parameter doesn't get options")
		}
	})
	callGood(theFrame, nil, Options{"foo": "bar"})

	// Combination of Frame and Options.
	f = NewBuiltinFn("f", func(f *Frame, opts Options) {
		if f != theFrame {
			t.Errorf("*Frame parameter doesn't get current frame")
		}
		if opts["foo"] != "bar" {
			t.Errorf("Options parameter doesn't get options")
		}
	})
	callGood(theFrame, nil, Options{"foo": "bar"})

	// Argument passing.
	f = NewBuiltinFn("f", func(x, y string) {
		if x != "lorem" {
			t.Errorf("Argument x not passed")
		}
		if y != "ipsum" {
			t.Errorf("Argument y not passed")
		}
	})
	callGood(theFrame, []interface{}{"lorem", "ipsum"}, theOptions)

	// Variadic arguments.
	f = NewBuiltinFn("f", func(x ...string) {
		if len(x) != 2 || x[0] != "lorem" || x[1] != "ipsum" {
			t.Errorf("Variadic argument not passed")
		}
	})
	callGood(theFrame, []interface{}{"lorem", "ipsum"}, theOptions)

	// Conversion into int and float64.
	f = NewBuiltinFn("f", func(i int, f float64) {
		if i != 314 {
			t.Errorf("Integer argument i not passed")
		}
		if f != 1.25 {
			t.Errorf("Float argument f not passed")
		}
	})
	callGood(theFrame, []interface{}{"314", "1.25"}, theOptions)

	// Conversion of supplied inputs.
	f = NewBuiltinFn("f", func(i Inputs) {
		var values []interface{}
		i(func(x interface{}) {
			values = append(values, x)
		})
		if len(values) != 2 || values[0] != "foo" || values[1] != "bar" {
			t.Errorf("Inputs parameter didn't get supplied inputs")
		}
	})
	callGood(theFrame, []interface{}{vals.MakeList("foo", "bar")}, theOptions)

	// Conversion of implicit inputs.
	inFrame := &Frame{ports: make([]*Port, 3)}
	ch := make(chan interface{}, 10)
	ch <- "foo"
	ch <- "bar"
	close(ch)
	inFrame.ports[0] = &Port{Chan: ch}
	f = NewBuiltinFn("f", func(i Inputs) {
		var values []interface{}
		i(func(x interface{}) {
			values = append(values, x)
		})
		if len(values) != 2 || values[0] != "foo" || values[1] != "bar" {
			t.Errorf("Inputs parameter didn't get implicit inputs")
		}
	})
	callGood(inFrame, []interface{}{vals.MakeList("foo", "bar")}, theOptions)

	// Outputting of return values.
	outFrame := &Frame{ports: make([]*Port, 3)}
	ch = make(chan interface{}, 10)
	outFrame.ports[1] = &Port{Chan: ch}
	f = NewBuiltinFn("f", func() string { return "ret" })
	callGood(outFrame, nil, theOptions)
	select {
	case ret := <-ch:
		if ret != "ret" {
			t.Errorf("Output is not the same as return value")
		}
	default:
		t.Errorf("Return value is not outputted")
	}

	// Conversion of return values.
	f = NewBuiltinFn("f", func() int { return 314 })
	callGood(outFrame, nil, theOptions)
	select {
	case ret := <-ch:
		if ret != "314" {
			t.Errorf("Return value is not converted to string")
		}
	default:
		t.Errorf("Return value is not outputted")
	}

	// Passing of error return value.
	theError := errors.New("the error")
	f = NewBuiltinFn("f", func() (string, error) {
		return "x", theError
	})
	if f.Call(outFrame, nil, theOptions) != theError {
		t.Errorf("Returned error is not passed")
	}
	select {
	case <-ch:
		t.Errorf("Return value is outputted when error is not nil")
	default:
	}

	// Too many arguments.
	f = NewBuiltinFn("f", func() {
		t.Errorf("Function called when there are too many arguments")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)

	// Too few arguments.
	f = NewBuiltinFn("f", func(x string) {
		t.Errorf("Function called when there are too few arguments")
	})
	callBad(theFrame, nil, theOptions)
	f = NewBuiltinFn("f", func(x string, y ...string) {
		t.Errorf("Function called when there are too few arguments")
	})
	callBad(theFrame, nil, theOptions)

	// Options when the function does not accept options.
	f = NewBuiltinFn("f", func() {
		t.Errorf("Function called when there are extra options")
	})
	callBad(theFrame, nil, Options{"foo": "bar"})

	// Wrong argument type.
	f = NewBuiltinFn("f", func(x string) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{1}, theOptions)

	// Wrong argument type: cannot convert to int.
	f = NewBuiltinFn("f", func(x int) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)

	// Wrong argument type: cannot convert to float64.
	f = NewBuiltinFn("f", func(x float64) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)
}
