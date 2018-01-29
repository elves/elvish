package eval

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
)

// String operations.

var ErrInput = errors.New("input error")

func init() {
	addToBuiltinFns([]*BuiltinFn{
		{"<s",
			wrapStrCompare(func(a, b string) bool { return a < b })},
		{"<=s",
			wrapStrCompare(func(a, b string) bool { return a <= b })},
		{"==s",
			wrapStrCompare(func(a, b string) bool { return a == b })},
		{"!=s",
			wrapStrCompare(func(a, b string) bool { return a != b })},
		{">s",
			wrapStrCompare(func(a, b string) bool { return a > b })},
		{">=s",
			wrapStrCompare(func(a, b string) bool { return a >= b })},

		{"to-string", toString},

		{"joins", joins},
		{"splits", splits},
		{"replaces", replaces},

		{"ord", ord},
		{"base", base},

		{"wcswidth", wcswidth},
		{"-override-wcwidth", overrideWcwidth},

		{"has-prefix", hasPrefix},
		{"has-suffix", hasSuffix},

		{"eawk", eawk},
	})
}

func wrapStrCompare(cmp func(a, b string) bool) BuiltinFnImpl {
	return func(ec *Frame, args []interface{}, opts map[string]interface{}) {
		TakeNoOpt(opts)
		for _, a := range args {
			if _, ok := a.(string); !ok {
				throw(ErrArgs)
			}
		}
		result := true
		for i := 0; i < len(args)-1; i++ {
			if !cmp(args[i].(string), args[i+1].(string)) {
				result = false
				break
			}
		}
		ec.OutputChan() <- types.Bool(result)
	}
}

// toString converts all arguments to strings.
func toString(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	out := ec.OutputChan()
	for _, a := range args {
		out <- types.ToString(a)
	}
}

// joins joins all input strings with a delimiter.
func joins(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var sepv string
	iterate := ScanArgsOptionalInput(ec, args, &sepv)
	sep := sepv
	TakeNoOpt(opts)

	var buf bytes.Buffer
	iterate(func(v interface{}) {
		if s, ok := v.(string); ok {
			if buf.Len() > 0 {
				buf.WriteString(sep)
			}
			buf.WriteString(s)
		} else {
			throwf("join wants string input, got %s", types.Kind(v))
		}
	})
	out := ec.ports[1].Chan
	out <- buf.String()
}

// splits splits an argument strings by a delimiter and writes all pieces.
func splits(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var (
		s, sep string
		optMax int
	)
	ScanArgs(args, &sep, &s)
	ScanOpts(opts, OptToScan{"max", &optMax, "-1"})

	out := ec.ports[1].Chan
	parts := strings.SplitN(s, sep, optMax)
	for _, p := range parts {
		out <- p
	}
}

func replaces(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var (
		old, repl, s string
		optMax       int
	)
	ScanArgs(args, &old, &repl, &s)
	ScanOpts(opts, OptToScan{"max", &optMax, "-1"})

	ec.ports[1].Chan <- strings.Replace(s, old, repl, optMax)
}

func ord(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var s string
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	for _, r := range s {
		out <- "0x" + strconv.FormatInt(int64(r), 16)
	}
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var (
		b    int
		nums []int
	)
	ScanArgsVariadic(args, &b, &nums)
	TakeNoOpt(opts)

	if b < 2 || b > 36 {
		throw(ErrBadBase)
	}

	out := ec.ports[1].Chan

	for _, num := range nums {
		out <- strconv.FormatInt(int64(num), b)
	}
}

func wcswidth(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var s string
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- strconv.Itoa(util.Wcswidth(s))
}

func overrideWcwidth(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var (
		s string
		w int
	)
	ScanArgs(args, &s, &w)
	TakeNoOpt(opts)

	r, err := toRune(s)
	maybeThrow(err)
	util.OverrideWcwidth(r, w)
}

func hasPrefix(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var s, prefix string
	ScanArgs(args, &s, &prefix)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(strings.HasPrefix(s, prefix))
}

func hasSuffix(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var s, suffix string
	ScanArgs(args, &s, &suffix)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(strings.HasSuffix(s, suffix))
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
func eawk(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var f Fn
	iterate := ScanArgsOptionalInput(ec, args, &f)
	TakeNoOpt(opts)

	broken := false
	iterate(func(v interface{}) {
		if broken {
			return
		}
		line, ok := v.(string)
		if !ok {
			throw(ErrInput)
		}
		args := []interface{}{line}
		for _, field := range eawkWordSep.Split(strings.Trim(line, " \t"), -1) {
			args = append(args, field)
		}

		newec := ec.fork("fn of eawk")
		// TODO: Close port 0 of newec.
		ex := newec.PCall(f, args, NoOpts)
		ClosePorts(newec.ports)

		if ex != nil {
			switch ex.(*Exception).Cause {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				throw(ex)
			}
		}
	})
}
