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
	return func(ec *Frame, args []types.Value, opts map[string]types.Value) {
		TakeNoOpt(opts)
		for _, a := range args {
			if _, ok := a.(types.String); !ok {
				throw(ErrArgs)
			}
		}
		result := true
		for i := 0; i < len(args)-1; i++ {
			if !cmp(string(args[i].(types.String)), string(args[i+1].(types.String))) {
				result = false
				break
			}
		}
		ec.OutputChan() <- types.Bool(result)
	}
}

// toString converts all arguments to strings.
func toString(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)
	out := ec.OutputChan()
	for _, a := range args {
		out <- types.String(types.ToString(a))
	}
}

// joins joins all input strings with a delimiter.
func joins(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var sepv types.String
	iterate := ScanArgsOptionalInput(ec, args, &sepv)
	sep := string(sepv)
	TakeNoOpt(opts)

	var buf bytes.Buffer
	iterate(func(v types.Value) {
		if s, ok := v.(types.String); ok {
			if buf.Len() > 0 {
				buf.WriteString(sep)
			}
			buf.WriteString(string(s))
		} else {
			throwf("join wants string input, got %s", v.Kind())
		}
	})
	out := ec.ports[1].Chan
	out <- types.String(buf.String())
}

// splits splits an argument strings by a delimiter and writes all pieces.
func splits(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var (
		s, sep types.String
		optMax int
	)
	ScanArgs(args, &sep, &s)
	ScanOpts(opts, OptToScan{"max", &optMax, types.String("-1")})

	out := ec.ports[1].Chan
	parts := strings.SplitN(string(s), string(sep), optMax)
	for _, p := range parts {
		out <- types.String(p)
	}
}

func replaces(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var (
		old, repl, s types.String
		optMax       int
	)
	ScanArgs(args, &old, &repl, &s)
	ScanOpts(opts, OptToScan{"max", &optMax, types.String("-1")})

	ec.ports[1].Chan <- types.String(strings.Replace(string(s), string(old), string(repl), optMax))
}

func ord(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var s types.String
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	for _, r := range s {
		out <- types.String("0x" + strconv.FormatInt(int64(r), 16))
	}
}

// ErrBadBase is thrown by the "base" builtin if the base is smaller than 2 or
// greater than 36.
var ErrBadBase = errors.New("bad base")

func base(ec *Frame, args []types.Value, opts map[string]types.Value) {
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
		out <- types.String(strconv.FormatInt(int64(num), b))
	}
}

func wcswidth(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var s types.String
	ScanArgs(args, &s)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- types.String(strconv.Itoa(util.Wcswidth(string(s))))
}

func overrideWcwidth(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var (
		s types.String
		w int
	)
	ScanArgs(args, &s, &w)
	TakeNoOpt(opts)

	r, err := toRune(s)
	maybeThrow(err)
	util.OverrideWcwidth(r, w)
}

func hasPrefix(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var s, prefix types.String
	ScanArgs(args, &s, &prefix)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(strings.HasPrefix(string(s), string(prefix)))
}

func hasSuffix(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var s, suffix types.String
	ScanArgs(args, &s, &suffix)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(strings.HasSuffix(string(s), string(suffix)))
}

var eawkWordSep = regexp.MustCompile("[ \t]+")

// eawk takes a function. For each line in the input stream, it calls the
// function with the line and the words in the line. The words are found by
// stripping the line and splitting the line by whitespaces. The function may
// call break and continue. Overall this provides a similar functionality to
// awk, hence the name.
func eawk(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var f Fn
	iterate := ScanArgsOptionalInput(ec, args, &f)
	TakeNoOpt(opts)

	broken := false
	iterate(func(v types.Value) {
		if broken {
			return
		}
		line, ok := v.(types.String)
		if !ok {
			throw(ErrInput)
		}
		args := []types.Value{line}
		for _, field := range eawkWordSep.Split(strings.Trim(string(line), " \t"), -1) {
			args = append(args, types.String(field))
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
