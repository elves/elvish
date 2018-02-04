package eval

import (
	"errors"
	"testing"
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
	f = NewReflectBuiltinFn("f", func(f *Frame) {
		if f != theFrame {
			t.Errorf("*Frame parameter doesn't get current frame")
		}
	})
	callGood(theFrame, nil, theOptions)

	// Options parameter gets options.
	f = NewReflectBuiltinFn("f", func(opts Options) {
		if opts["foo"] != "bar" {
			t.Errorf("Options parameter doesn't get options")
		}
	})
	callGood(theFrame, nil, Options{"foo": "bar"})

	// Combination of Frame and Options.
	f = NewReflectBuiltinFn("f", func(f *Frame, opts Options) {
		if f != theFrame {
			t.Errorf("*Frame parameter doesn't get current frame")
		}
		if opts["foo"] != "bar" {
			t.Errorf("Options parameter doesn't get options")
		}
	})
	callGood(theFrame, nil, Options{"foo": "bar"})

	// Argument passing.
	f = NewReflectBuiltinFn("f", func(x, y string) {
		if x != "lorem" {
			t.Errorf("Argument x not passed")
		}
		if y != "ipsum" {
			t.Errorf("Argument y not passed")
		}
	})
	callGood(theFrame, []interface{}{"lorem", "ipsum"}, theOptions)

	// Variadic arguments.
	f = NewReflectBuiltinFn("f", func(x ...string) {
		if len(x) != 2 || x[0] != "lorem" || x[1] != "ipsum" {
			t.Errorf("Variadic argument not passed")
		}
	})
	callGood(theFrame, []interface{}{"lorem", "ipsum"}, theOptions)

	// Conversion into int and float64.
	f = NewReflectBuiltinFn("f", func(i int, f float64) {
		if i != 314 {
			t.Errorf("Integer argument i not passed")
		}
		if f != 1.25 {
			t.Errorf("Float argument f not passed")
		}
	})
	callGood(theFrame, []interface{}{"314", "1.25"}, theOptions)

	// Outputting of return values.
	outFrame := &Frame{ports: make([]*Port, 3)}
	ch := make(chan interface{}, 10)
	outFrame.ports[1] = &Port{Chan: ch}
	f = NewReflectBuiltinFn("f", func() string { return "ret" })
	callGood(outFrame, nil, theOptions)
	if <-ch != "ret" {
		t.Errorf("Return value not outputted")
	}

	// Passing of error return value.
	theError := errors.New("the error")
	f = NewReflectBuiltinFn("f", func() error { return theError })
	if f.Call(theFrame, nil, theOptions) != theError {
		t.Errorf("Returned error is not passed")
	}

	// Too many arguments.
	f = NewReflectBuiltinFn("f", func() {
		t.Errorf("Function called when there are too many arguments")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)

	// Too few arguments.
	f = NewReflectBuiltinFn("f", func(x string) {
		t.Errorf("Function called when there are too few arguments")
	})
	callBad(theFrame, nil, theOptions)
	f = NewReflectBuiltinFn("f", func(x string, y ...string) {
		t.Errorf("Function called when there are too few arguments")
	})
	callBad(theFrame, nil, theOptions)

	// Wrong argument type.
	f = NewReflectBuiltinFn("f", func(x string) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{1}, theOptions)

	// Wrong argument type: cannot convert to int.
	f = NewReflectBuiltinFn("f", func(x int) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)

	// Wrong argument type: cannot convert to float64.
	f = NewReflectBuiltinFn("f", func(x float64) {
		t.Errorf("Function called when arguments have wrong type")
	})
	callBad(theFrame, []interface{}{"x"}, theOptions)
}
