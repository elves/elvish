package re

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

func Namespace() eval.Namespace {
	ns := eval.Namespace{}
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

func match(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	out := ec.OutputChan()
	var (
		argPattern eval.String
		argSource  eval.String
		optPOSIX   eval.Bool
	)
	eval.ScanArgs(args, &argPattern, &argSource)
	eval.ScanOpts(opts, eval.Opt{"posix", &optPOSIX, eval.Bool(false)})

	pattern := makePattern(argPattern, optPOSIX, eval.Bool(false))
	matched := pattern.MatchString(string(argSource))
	out <- eval.Bool(matched)
}

var (
	matchFields    = []string{"text", "start", "end", "groups"}
	submatchFields = []string{"text", "start", "end"}
)

func find(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	out := ec.OutputChan()
	var (
		argPattern eval.String
		argSource  eval.String
		optPOSIX   eval.Bool
		optLongest eval.Bool
		optMax     int
	)
	eval.ScanArgs(args, &argPattern, &argSource)
	eval.ScanOpts(opts,
		eval.Opt{"posix", &optPOSIX, eval.Bool(false)},
		eval.Opt{"longest", &optLongest, eval.Bool(false)},
		eval.Opt{"max", &optMax, eval.String("-1")})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	matches := pattern.FindAllSubmatchIndex([]byte(argSource), optMax)
	for _, match := range matches {
		start, end := match[0], match[1]
		groups := vector.Empty
		for i := 0; i < len(match); i += 2 {
			start, end := match[i], match[i+1]
			groups = groups.Cons(&eval.Struct{submatchFields, []eval.Variable{
				eval.NewRoVariable(argSource[start:end]),
				eval.NewRoVariable(eval.String(strconv.Itoa(start))),
				eval.NewRoVariable(eval.String(strconv.Itoa(end))),
			}})
		}
		out <- &eval.Struct{matchFields, []eval.Variable{
			eval.NewRoVariable(argSource[start:end]),
			eval.NewRoVariable(eval.String(strconv.Itoa(start))),
			eval.NewRoVariable(eval.String(strconv.Itoa(end))),
			eval.NewRoVariable(eval.NewListFromVector(groups)),
		}}
	}
}

func replace(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	out := ec.OutputChan()
	var (
		argPattern eval.String
		argRepl    eval.Value
		argSource  eval.String
		optPOSIX   eval.Bool
		optLongest eval.Bool
		optLiteral eval.Bool
	)
	eval.ScanArgs(args, &argPattern, &argRepl, &argSource)
	eval.ScanOpts(opts,
		eval.Opt{"posix", &optPOSIX, eval.Bool(false)},
		eval.Opt{"longest", &optLongest, eval.Bool(false)},
		eval.Opt{"literal", &optLiteral, eval.Bool(false)})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	var result string
	if optLiteral {
		repl, ok := argRepl.(eval.String)
		if !ok {
			throwf("replacement must be string when literal is set, got %s",
				argRepl.Kind())
		}
		result = pattern.ReplaceAllLiteralString(string(argSource), string(repl))
	} else {
		switch repl := argRepl.(type) {
		case eval.String:
			result = pattern.ReplaceAllString(string(argSource), string(repl))
		case eval.CallableValue:
			replFunc := func(s string) string {
				values, err := ec.PCaptureOutput(repl,
					[]eval.Value{eval.String(s)}, eval.NoOpts)
				maybeThrow(err)
				if len(values) != 1 {
					throwf("replacement function must output exactly one value, got %d", len(values))
				}
				output, ok := values[0].(eval.String)
				if !ok {
					throwf("replacement function must output one string, got %s", values[0].Kind())
				}
				return string(output)
			}
			result = pattern.ReplaceAllStringFunc(string(argSource), replFunc)
		default:
			throwf("replacement must be string or function, got %s",
				argRepl.Kind())
		}
	}
	out <- eval.String(result)
}

func split(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	out := ec.OutputChan()
	var (
		argPattern eval.String
		argSource  eval.String
		optPOSIX   eval.Bool
		optLongest eval.Bool
		optMax     int
	)
	eval.ScanArgs(args, &argPattern, &argSource)
	eval.ScanOpts(opts,
		eval.Opt{"posix", &optPOSIX, eval.Bool(false)},
		eval.Opt{"longest", &optLongest, eval.Bool(false)},
		eval.Opt{"max", &optMax, eval.String("-1")})

	pattern := makePattern(argPattern, optPOSIX, optLongest)

	pieces := pattern.Split(string(argSource), optMax)
	for _, piece := range pieces {
		out <- eval.String(piece)
	}
}

func makePattern(argPattern eval.String,
	optPOSIX, optLongest eval.Bool) *regexp.Regexp {

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
