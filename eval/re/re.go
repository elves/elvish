// Package re implements the re: module for using regular expressions.
package re

import (
	"fmt"
	"regexp"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

func Ns() eval.Ns {
	ns := eval.Ns{}
	eval.AddBuiltinFns(ns, fns...)
	return ns
}

var fns = []*eval.BuiltinFn{
	{"quote", eval.WrapStringToString(regexp.QuoteMeta)},
	{"match", match},
	{"find", find},
	{"replace", replace},
	{"split", split},
}

func match(ec *eval.Frame, args []types.Value, opts map[string]types.Value) {
	out := ec.OutputChan()
	var (
		argPattern types.String
		argSource  types.String
		optPOSIX   types.Bool
	)
	eval.ScanArgs(args, &argPattern, &argSource)
	eval.ScanOpts(opts, eval.OptToScan{"posix", &optPOSIX, types.Bool(false)})

	pattern := makePattern(argPattern, optPOSIX, types.Bool(false))
	matched := pattern.MatchString(string(argSource))
	out <- types.Bool(matched)
}

func find(ec *eval.Frame, args []types.Value, opts map[string]types.Value) {
	out := ec.OutputChan()
	var (
		argPattern types.String
		argSource  types.String
		optPOSIX   types.Bool
		optLongest types.Bool
		optMax     int
	)
	eval.ScanArgs(args, &argPattern, &argSource)
	eval.ScanOpts(opts,
		eval.OptToScan{"posix", &optPOSIX, types.Bool(false)},
		eval.OptToScan{"longest", &optLongest, types.Bool(false)},
		eval.OptToScan{"max", &optMax, types.String("-1")})

	pattern := makePattern(argPattern, optPOSIX, optLongest)
	source := string(argSource)

	matches := pattern.FindAllSubmatchIndex([]byte(argSource), optMax)
	for _, match := range matches {
		start, end := match[0], match[1]
		groups := vector.Empty
		for i := 0; i < len(match); i += 2 {
			start, end := match[i], match[i+1]
			text := ""
			// FindAllSubmatchIndex may return negative indicies to indicate
			// that the pattern didn't appear in the text.
			if start >= 0 && end >= 0 {
				text = source[start:end]
			}
			groups = groups.Cons(newSubmatch(text, start, end))
		}
		out <- newMatch(source[start:end], start, end, groups)
	}
}

func replace(ec *eval.Frame, args []types.Value, opts map[string]types.Value) {
	out := ec.OutputChan()
	var (
		argPattern types.String
		argRepl    types.Value
		argSource  types.String
		optPOSIX   types.Bool
		optLongest types.Bool
		optLiteral types.Bool
	)
	eval.ScanArgs(args, &argPattern, &argRepl, &argSource)
	eval.ScanOpts(opts,
		eval.OptToScan{"posix", &optPOSIX, types.Bool(false)},
		eval.OptToScan{"longest", &optLongest, types.Bool(false)},
		eval.OptToScan{"literal", &optLiteral, types.Bool(false)})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	var result string
	if optLiteral {
		repl, ok := argRepl.(types.String)
		if !ok {
			throwf("replacement must be string when literal is set, got %s",
				types.Kind(argRepl))
		}
		result = pattern.ReplaceAllLiteralString(string(argSource), string(repl))
	} else {
		switch repl := argRepl.(type) {
		case types.String:
			result = pattern.ReplaceAllString(string(argSource), string(repl))
		case eval.Fn:
			replFunc := func(s string) string {
				values, err := ec.PCaptureOutput(repl,
					[]types.Value{types.String(s)}, eval.NoOpts)
				maybeThrow(err)
				if len(values) != 1 {
					throwf("replacement function must output exactly one value, got %d", len(values))
				}
				output, ok := values[0].(types.String)
				if !ok {
					throwf("replacement function must output one string, got %s",
						types.Kind(values[0]))
				}
				return string(output)
			}
			result = pattern.ReplaceAllStringFunc(string(argSource), replFunc)
		default:
			throwf("replacement must be string or function, got %s",
				types.Kind(argRepl))
		}
	}
	out <- types.String(result)
}

func split(ec *eval.Frame, args []types.Value, opts map[string]types.Value) {
	out := ec.OutputChan()
	var (
		argPattern types.String
		argSource  types.String
		optPOSIX   types.Bool
		optLongest types.Bool
		optMax     int
	)
	eval.ScanArgs(args, &argPattern, &argSource)
	eval.ScanOpts(opts,
		eval.OptToScan{"posix", &optPOSIX, types.Bool(false)},
		eval.OptToScan{"longest", &optLongest, types.Bool(false)},
		eval.OptToScan{"max", &optMax, types.String("-1")})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	pieces := pattern.Split(string(argSource), optMax)
	for _, piece := range pieces {
		out <- types.String(piece)
	}
}

func makePattern(argPattern types.String, optPOSIX, optLongest types.Bool) *regexp.Regexp {
	var (
		pattern *regexp.Regexp
		err     error
	)
	if optPOSIX {
		pattern, err = regexp.CompilePOSIX(string(argPattern))
	} else {
		pattern, err = regexp.Compile(string(argPattern))
	}
	maybeThrow(err)
	if optLongest {
		pattern.Longest()
	}
	return pattern
}

func throwf(format string, args ...interface{}) {
	util.Throw(fmt.Errorf(format, args...))
}

func maybeThrow(err error) {
	if err != nil {
		util.Throw(err)
	}
}
