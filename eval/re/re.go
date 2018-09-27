// Package re implements the re: module for using regular expressions.
package re

import (
	"fmt"
	"regexp"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
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

func match(rawOpts eval.RawOptions, argPattern, source string) (bool, error) {
	opts := struct{ Posix bool }{}
	rawOpts.Scan(&opts)

	pattern, err := makePattern(argPattern, opts.Posix, false)
	if err != nil {
		return false, err
	}
	return pattern.MatchString(source), nil
}

func find(fm *eval.Frame, rawOpts eval.RawOptions, argPattern, source string) error {
	out := fm.OutputChan()
	opts := struct {
		Posix   bool
		Longest bool
		Max     int
	}{Max: -1}
	rawOpts.Scan(&opts)

	pattern, err := makePattern(argPattern, opts.Posix, opts.Longest)
	if err != nil {
		return err
	}
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
	return nil
}

func replace(fm *eval.Frame, rawOpts eval.RawOptions, argPattern string, argRepl interface{}, source string) (string, error) {

	opts := struct {
		Posix   bool
		Longest bool
		Literal bool
	}{}
	rawOpts.Scan(&opts)

	pattern, err := makePattern(argPattern, opts.Posix, opts.Longest)
	if err != nil {
		return "", err
	}

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
			var errReplace error
			replFunc := func(s string) string {
				if errReplace != nil {
					return ""
				}
				values, err := fm.CaptureOutput(repl, []interface{}{s}, eval.NoOpts)
				if err != nil {
					errReplace = err
					return ""
				}
				if len(values) != 1 {
					errReplace = fmt.Errorf("replacement function must output exactly one value, got %d", len(values))
					return ""
				}
				output, ok := values[0].(string)
				if !ok {
					errReplace = fmt.Errorf(
						"replacement function must output one string, got %s",
						vals.Kind(values[0]))
					return ""
				}
				return output
			}
			return pattern.ReplaceAllStringFunc(source, replFunc), errReplace
		default:
			return "", fmt.Errorf(
				"replacement must be string or function, got %s",
				vals.Kind(argRepl))
		}
	}
}

func split(fm *eval.Frame, rawOpts eval.RawOptions, argPattern, source string) error {
	out := fm.OutputChan()
	opts := struct {
		Posix   bool
		Longest bool
		Max     int
	}{Max: -1}
	rawOpts.Scan(&opts)

	pattern, err := makePattern(argPattern, opts.Posix, opts.Longest)
	if err != nil {
		return err
	}

	pieces := pattern.Split(source, opts.Max)
	for _, piece := range pieces {
		out <- piece
	}
	return nil
}

func makePattern(p string, posix, longest bool) (*regexp.Regexp, error) {
	pattern, err := compile(p, posix)
	if err != nil {
		return nil, err
	}
	if longest {
		pattern.Longest()
	}
	return pattern, nil
}

func compile(pattern string, posix bool) (*regexp.Regexp, error) {
	if posix {
		return regexp.CompilePOSIX(pattern)
	}
	return regexp.Compile(pattern)
}
