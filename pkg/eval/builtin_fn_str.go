package eval

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/wcwidth"
)

// String operations.

// ErrInputOfEawkMustBeString is thrown when eawk gets a non-string input.
var ErrInputOfEawkMustBeString = errors.New("input of eawk must be string")

// TODO(xiaq): Document -override-wcswidth.

func init() {
	addBuiltinFns(map[string]any{
		"<s":  func(a, b string) bool { return a < b },
		"<=s": func(a, b string) bool { return a <= b },
		"==s": func(a, b string) bool { return a == b },
		"!=s": func(a, b string) bool { return a != b },
		">s":  func(a, b string) bool { return a > b },
		">=s": func(a, b string) bool { return a >= b },

		"to-nums": toNums,
		"from-nums": fromNums,

		"to-string": toString,

		"base": base,

		"wcswidth":          wcwidth.Of,
		"-override-wcwidth": wcwidth.Override,

		"eawk": eawk,
	})
}

func toNums(fm *Frame, args ...any) error {
	in := fm.InputFile()
	out := fm.ValueOutput()
	for true {
		c := make([]byte, 1)
		n, err := in.Read(c)
		if(n != 1 || err != nil) {
			break
		}
		err = out.Put(int(c[0]))
		if(err != nil) {
			return err
		}
	}
	return nil
}

func numToByte(v any) (byte, error) {
	switch v := v.(type) {
	case int:
		if(v > 255) {
			return 0, errors.New("must be less than 255");
		}
		return byte(v), nil
	case float64:
		if(v > 255) {
			return 0, errors.New("must be less than 255");
		}
		return byte(v), nil
	default:
		return 0, errors.New("must be number")
	}
}

func fromNums(fm *Frame, args ...any) error {

	in := fm.InputChan()
	out := fm.ByteOutput()
	for true {
		v, ok := <- in;
		if(!ok) {
			break
		}
		t, err := numToByte(v)
		if(err != nil) {
			return err
		}
		c := make([]byte, 1)
		c[0] = t
		n, err := out.Write(c)

		if(err != nil || n != 1) {
			return err
		}
	}
	return nil
}


func toString(fm *Frame, args ...any) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(vals.ToString(a))
		if err != nil {
			return err
		}
	}
	return nil
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(fm *Frame, b int, nums ...int) error {
	if b < 2 || b > 36 {
		return ErrBadBase
	}

	out := fm.ValueOutput()
	for _, num := range nums {
		err := out.Put(strconv.FormatInt(int64(num), b))
		if err != nil {
			return err
		}
	}
	return nil
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

func eawk(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v any) {
		if broken {
			return
		}
		line, ok := v.(string)
		if !ok {
			broken = true
			err = ErrInputOfEawkMustBeString
			return
		}
		args := []any{line}
		for _, field := range eawkWordSep.Split(strings.Trim(line, " \t"), -1) {
			args = append(args, field)
		}

		newFm := fm.Fork("fn of eawk")
		// TODO: Close port 0 of newFm.
		ex := f.Call(newFm, args, NoOpts)
		newFm.Close()

		if ex != nil {
			switch Reason(ex) {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				broken = true
				err = ex
			}
		}
	})
	return err
}
