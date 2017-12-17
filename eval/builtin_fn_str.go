package eval

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"

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
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		for _, a := range args {
			if _, ok := a.(String); !ok {
				throw(ErrArgs)
			}
		}
		result := true
		for i := 0; i < len(args)-1; i++ {
			if !cmp(string(args[i].(String)), string(args[i+1].(String))) {
				result = false
				break
			}
		}
		ec.OutputChan() <- Bool(result)
	}
}

// toString converts all arguments to strings.
func toString(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)
	out := ec.OutputChan()
	for _, a := range args {
		out <- String(ToString(a))
	}
}

// joins joins all input strings with a delimiter.
func joins(ec *EvalCtx, args []Value, opts map[string]Value) {
	var sepv String
	iterate := ScanArgsOptionalInput(ec, args, &sepv)
	sep := string(sepv)
	TakeNoOpt(opts)

	var buf bytes.Buffer
	iterate(func(v Value) {
		if s, ok := v.(String); ok {
			if buf.Len() > 0 {
				buf.WriteString(sep)
			}
			buf.WriteString(string(s))
		} else {
			throwf("join wants string input, got %s", v.Kind())
		}
	})
	out := ec.ports[1].Chan
	out <- String(buf.String())
}

// splits splits an argument strings by a delimiter and writes all pieces.
func splits(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s, sep String
	ScanArgs(args, &sep, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	parts := strings.Split(string(s), string(sep))
	for _, p := range parts {
		out <- String(p)
	}
}

func replaces(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		old, repl, s String
		optMax       int
	)
	ScanArgs(args, &old, &repl, &s)
	ScanOpts(opts, OptToScan{"max", &optMax, String("-1")})

	ec.ports[1].Chan <- String(strings.Replace(string(s), string(old), string(repl), optMax))
}

func ord(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s String
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	for _, r := range s {
		out <- String("0x" + strconv.FormatInt(int64(r), 16))
	}
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(ec *EvalCtx, args []Value, opts map[string]Value) {
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
		out <- String(strconv.FormatInt(int64(num), b))
	}
}

func wcswidth(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s String
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(strconv.Itoa(util.Wcswidth(string(s))))
}

func overrideWcwidth(ec *EvalCtx, args []Value, opts map[string]Value) {
	var (
		s String
		w int
	)
	ScanArgs(args, &s, &w)
	TakeNoOpt(opts)

	r, err := toRune(s)
	maybeThrow(err)
	util.OverrideWcwidth(r, w)
}

func hasPrefix(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s, prefix String
	ScanArgs(args, &s, &prefix)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(strings.HasPrefix(string(s), string(prefix)))
}

func hasSuffix(ec *EvalCtx, args []Value, opts map[string]Value) {
	var s, suffix String
	ScanArgs(args, &s, &suffix)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(strings.HasSuffix(string(s), string(suffix)))
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
func eawk(ec *EvalCtx, args []Value, opts map[string]Value) {
	var f CallableValue
	iterate := ScanArgsOptionalInput(ec, args, &f)
	TakeNoOpt(opts)

	broken := false
	iterate(func(v Value) {
		if broken {
			return
		}
		line, ok := v.(String)
		if !ok {
			throw(ErrInput)
		}
		args := []Value{line}
		for _, field := range eawkWordSep.Split(strings.Trim(string(line), " \t"), -1) {
			args = append(args, String(field))
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
