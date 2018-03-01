// Package re implements the re: module for using regular expressions.
package re

import (
	"fmt"
	"regexp"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

var Ns = eval.NewNs().AddBuiltinFns("re:", fns)

var fns = map[string]interface{}{
	"quote":   regexp.QuoteMeta,
	"match":   match,
	"find":    find,
	"replace": replace,
	"split":   split,
}

func match(rawOpts eval.RawOptions, argPattern, source string) bool {
	opts := struct{ Posix bool }{}
	rawOpts.Scan(&opts)

	pattern := makePattern(argPattern, opts.Posix, false)
	return pattern.MatchString(source)
}

func find(fm *eval.Frame, rawOpts eval.RawOptions, argPattern, source string) {
	out := fm.OutputChan()
	opts := struct {
		Posix   bool
		Longest bool
		Max     int
	}{Max: -1}
	rawOpts.Scan(&opts)

	pattern := makePattern(argPattern, opts.Posix, opts.Longest)
	matches := pattern.FindAllSubmatchIndex([]byte(source), opts.Max)

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

func replace(fm *eval.Frame, rawOpts eval.RawOptions, argPattern string, argRepl interface{}, source string) (string, error) {

	opts := struct {
		Posix   bool
		Longest bool
		Literal bool
	}{}
	rawOpts.Scan(&opts)

	pattern := makePattern(argPattern, opts.Posix, opts.Longest)

	if opts.Literal {
		repl, ok := argRepl.(string)
		if !ok {
			return "", fmt.Errorf(
				"replacement must be string when literal is set, got %s",
				vals.Kind(argRepl))
		}
		return pattern.ReplaceAllLiteralString(source, repl), nil
	} else {
		switch repl := argRepl.(type) {
		case string:
			return pattern.ReplaceAllString(source, repl), nil
		case eval.Callable:
			replFunc := func(s string) string {
				values, err := fm.CaptureOutput(repl, []interface{}{s}, eval.NoOpts)
				maybeThrow(err)
				if len(values) != 1 {
					throwf("replacement function must output exactly one value, got %d", len(values))
				}
				output, ok := values[0].(string)
				if !ok {
					throwf("replacement function must output one string, got %s",
						vals.Kind(values[0]))
				}
				return output
			}
			return pattern.ReplaceAllStringFunc(source, replFunc), nil
		default:
			return "", fmt.Errorf(
				"replacement must be string or function, got %s",
				vals.Kind(argRepl))
		}
	}
}

func split(fm *eval.Frame, rawOpts eval.RawOptions, argPattern, source string) {
	out := fm.OutputChan()
	opts := struct {
		Posix   bool
		Longest bool
		Max     int
	}{Max: -1}
	rawOpts.Scan(&opts)

	pattern := makePattern(argPattern, opts.Posix, opts.Longest)

	pieces := pattern.Split(source, opts.Max)
	for _, piece := range pieces {
		out <- piece
	}
}

func makePattern(argPattern string, optPOSIX, optLongest bool) *regexp.Regexp {
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
