package eval

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

// String operations.

var ErrInput = errors.New("input error")

func init() {
	addBuiltinFns(map[string]interface{}{
		"<s":  func(a, b string) bool { return a < b },
		"<=s": func(a, b string) bool { return a <= b },
		"==s": func(a, b string) bool { return a == b },
		"!=s": func(a, b string) bool { return a != b },
		">s":  func(a, b string) bool { return a > b },
		">=s": func(a, b string) bool { return a >= b },

		"to-string": toString,

		"ord":  ord,
		"base": base,

		"wcswidth":          util.Wcswidth,
		"-override-wcwidth": util.OverrideWcwidth,

		"has-prefix": strings.HasPrefix,
		"has-suffix": strings.HasSuffix,

		"joins":    joins,
		"splits":   splits,
		"replaces": replaces,

		"eawk": eawk,
	})
}

// toString converts all arguments to strings.
func toString(fm *Frame, args ...interface{}) {
	out := fm.OutputChan()
	for _, a := range args {
		out <- vals.ToString(a)
	}
}

// joins joins all input strings with a delimiter.
func joins(sep string, inputs Inputs) string {
	var buf bytes.Buffer
	first := true
	inputs(func(v interface{}) {
		if s, ok := v.(string); ok {
			if first {
				first = false
			} else {
				buf.WriteString(sep)
			}
			buf.WriteString(s)
		} else {
			throwf("join wants string input, got %s", vals.Kind(v))
		}
	})
	return buf.String()
}

// splits splits an argument strings by a delimiter and writes all pieces.
func splits(fm *Frame, rawOpts RawOptions, sep, s string) {
	opts := struct{ Max int }{-1}
	rawOpts.Scan(&opts)

	out := fm.ports[1].Chan
	parts := strings.SplitN(s, sep, opts.Max)
	for _, p := range parts {
		out <- p
	}
}

func replaces(rawOpts RawOptions, old, repl, s string) string {
	opts := struct{ Max int }{-1}
	rawOpts.Scan(&opts)
	return strings.Replace(s, old, repl, opts.Max)
}

func ord(fm *Frame, s string) {
	out := fm.ports[1].Chan
	for _, r := range s {
		out <- "0x" + strconv.FormatInt(int64(r), 16)
	}
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(fm *Frame, b int, nums ...int) error {
	if b < 2 || b > 36 {
		return ErrBadBase
	}

	out := fm.ports[1].Chan
	for _, num := range nums {
		out <- strconv.FormatInt(int64(num), b)
	}
	return nil
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
func eawk(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v interface{}) {
		if broken {
			return
		}
		line, ok := v.(string)
		if !ok {
			broken = true
			err = ErrInput
			return
		}
		args := []interface{}{line}
		for _, field := range eawkWordSep.Split(strings.Trim(line, " \t"), -1) {
			args = append(args, field)
		}

		newFm := fm.fork("fn of eawk")
		// TODO: Close port 0 of newFm.
		ex := newFm.Call(f, args, NoOpts)
		newFm.Close()

		if ex != nil {
			switch ex.(*Exception).Cause {
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
