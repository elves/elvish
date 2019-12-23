package eval

import (
	"errors"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
)

type testOptions struct {
	Foo string
	Bar string
}

func (o *testOptions) SetDefaultOptions() { o.Bar = "default" }

func TestGoFnCall(t *testing.T) {
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
	f = NewGoFn("f", func(f *Frame) {
		if f != theFrame {
			t.Errorf("*Frame parameter doesn't get current frame")
		}
	})
	callGood(theFrame, nil, theOptions)

	// RawOptions parameter gets options.
	f = NewGoFn("f", func(opts RawOptions) {
		if opts["foo"] != "bar" {
			t.Errorf("RawOptions parameter doesn't get options")
		}
	})
	callGood(theFrame, nil, RawOptions{"foo": "bar"})

	// ScanOptions parameters gets scanned options.
	f = NewGoFn("f", func(opts testOptions) {
		if opts.Foo != "bar" {
			t.Errorf("ScanOptions parameter doesn't get options")
		}
		if opts.Bar != "default" {
			t.Errorf("ScanOptions parameter doesn't use default value")
		}
	})
	callGood(theFrame, nil, RawOptions{"foo": "bar"})

	// Combination of Frame and RawOptions.
	f = NewGoFn("f", func(f *Frame, opts RawOptions) {
		if f != theFrame {
			t.Errorf("*Frame parameter doesn't get current frame")
		}
		if opts["foo"] != "bar" {
			t.Errorf("RawOptions parameter doesn't get options")
		}
	})
	callGood(theFrame, nil, RawOptions{"foo": "bar"})

	// Argument passing.
	f = NewGoFn("f", func(x, y string) {
		if x != "lorem" {
			t.Errorf("Argument x not passed")
		}
		if y != "ipsum" {
			t.Errorf("Argument y not passed")
		}
	})
	callGood(theFrame, []interface{}{"lorem", "ipsum"}, theOptions)

	// Variadic arguments.
	f = NewGoFn("f", func(x ...string) {
		if len(x) != 2 || x[0] != "lorem" || x[1] != "ipsum" {
			t.Errorf("Variadic argument not passed")
		}
	})
	callGood(theFrame, []interface{}{"lorem", "ipsum"}, theOptions)

	// Conversion into int and float64.
	f = NewGoFn("f", func(i int, f float64) {
		if i != 314 {
			t.Errorf("Integer argument i not passed")
		}
		if f != 1.25 {
			t.Errorf("Float argument f not passed")
		}
	})
	callGood(theFrame, []interface{}{"314", "1.25"}, theOptions)

	// Conversion of supplied inputs.
	f = NewGoFn("f", func(i Inputs) {
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
	f = NewGoFn("f", func(i Inputs) {
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
	f = NewGoFn("f", func() string { return "ret" })
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
	f = NewGoFn("f", func() int { return 314 })
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
	f = NewGoFn("f", func() (string, error) {
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
	f = NewGoFn("f", func() {
		t.Errorf("Function called when there are too many arguments")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)

	// Too few arguments.
	f = NewGoFn("f", func(x string) {
		t.Errorf("Function called when there are too few arguments")
	})
	callBad(theFrame, nil, theOptions)
	f = NewGoFn("f", func(x string, y ...string) {
		t.Errorf("Function called when there are too few arguments")
	})
	callBad(theFrame, nil, theOptions)

	// Options when the function does not accept options.
	f = NewGoFn("f", func() {
		t.Errorf("Function called when there are extra options")
	})
	callBad(theFrame, nil, RawOptions{"foo": "bar"})

	// Wrong argument type.
	f = NewGoFn("f", func(x string) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{1}, theOptions)

	// Wrong argument type: cannot convert to int.
	f = NewGoFn("f", func(x int) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)

	// Wrong argument type: cannot convert to float64.
	f = NewGoFn("f", func(x float64) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)
}
