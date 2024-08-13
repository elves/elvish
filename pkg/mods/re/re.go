// Package re implements a regular expression module.
package re

import (
	"errors"
	"regexp"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Ns is the namespace for the re: module.
var Ns = eval.BuildNsNamed("re").
	AddGoFns(map[string]any{
		"quote":   regexp.QuoteMeta,
		"match":   match,
		"find":    find,
		"replace": replace,
		"split":   split,
		"awk":     awk,
	}).Ns()

type matchOpts struct{ Posix bool }

func (*matchOpts) SetDefaultOptions() {}

func match(opts matchOpts, argPattern, source string) (bool, error) {
	pattern, err := makePattern(argPattern, opts.Posix, false)
	if err != nil {
		return false, err
	}
	return pattern.MatchString(source), nil
}

// Struct for holding options to find. Also used by split.
type findOpts struct {
	Posix   bool
	Longest bool
	Max     int
}

func (o *findOpts) SetDefaultOptions() { o.Max = -1 }

func find(fm *eval.Frame, opts findOpts, argPattern, source string) error {
	out := fm.ValueOutput()

	pattern, err := makePattern(argPattern, opts.Posix, opts.Longest)
	if err != nil {
		return err
	}
	matches := pattern.FindAllSubmatchIndex([]byte(source), opts.Max)

	for _, match := range matches {
		start, end := match[0], match[1]
		groups := vals.EmptyList
		for i := 0; i < len(match); i += 2 {
			start, end := match[i], match[i+1]
			text := ""
			// FindAllSubmatchIndex may return negative indices to indicate
			// that the pattern didn't appear in the text.
			if start >= 0 && end >= 0 {
				text = source[start:end]
			}
			groups = groups.Conj(submatchStruct{text, start, end})
		}
		err := out.Put(matchStruct{source[start:end], start, end, groups})
		if err != nil {
			return err
		}
	}
	return nil
}

type replaceOpts struct {
	Posix   bool
	Longest bool
	Literal bool
}

func (*replaceOpts) SetDefaultOptions() {}

func replace(fm *eval.Frame, opts replaceOpts, argPattern string, argRepl any, source string) (string, error) {

	pattern, err := makePattern(argPattern, opts.Posix, opts.Longest)
	if err != nil {
		return "", err
	}

	if opts.Literal {
		repl, ok := argRepl.(string)
		if !ok {
			return "", &errs.BadValue{What: "literal replacement",
				Valid: "string", Actual: vals.Kind(argRepl)}
		}
		return pattern.ReplaceAllLiteralString(source, repl), nil
	}
	switch repl := argRepl.(type) {
	case string:
		return pattern.ReplaceAllString(source, repl), nil
	case eval.Callable:
		var errReplace error
		replFunc := func(s string) string {
			if errReplace != nil {
				return ""
			}
			values, err := fm.CaptureOutput(func(fm *eval.Frame) error {
				return repl.Call(fm, []any{s}, eval.NoOpts)
			})
			if err != nil {
				errReplace = err
				return ""
			}
			if len(values) != 1 {
				errReplace = &errs.ArityMismatch{What: "replacement function output",
					ValidLow: 1, ValidHigh: 1, Actual: len(values)}
				return ""
			}
			output, ok := values[0].(string)
			if !ok {
				errReplace = &errs.BadValue{What: "replacement function output",
					Valid: "string", Actual: vals.Kind(values[0])}
				return ""
			}
			return output
		}
		return pattern.ReplaceAllStringFunc(source, replFunc), errReplace
	default:
		return "", &errs.BadValue{What: "replacement",
			Valid: "string or function", Actual: vals.Kind(argRepl)}
	}
}

func split(fm *eval.Frame, opts findOpts, argPattern, source string) error {
	out := fm.ValueOutput()

	pattern, err := makePattern(argPattern, opts.Posix, opts.Longest)
	if err != nil {
		return err
	}

	pieces := pattern.Split(source, opts.Max)
	for _, piece := range pieces {
		err := out.Put(piece)
		if err != nil {
			return err
		}
	}
	return nil
}

// ErrInputOfAwkMustBeString is thrown when re:awk gets a non-string input.
var ErrInputOfAwkMustBeString = errors.New("input of re:awk must be string")

type awkOpt struct {
	Sep        string
	SepPosix   bool
	SepLongest bool
}

func (o *awkOpt) SetDefaultOptions() {
	o.Sep = "[ \t]+"
}

func awk(fm *eval.Frame, opts awkOpt, f eval.Callable, inputs eval.Inputs) error {
	wordSep, err := makePattern(opts.Sep, opts.SepPosix, opts.SepLongest)
	if err != nil {
		return err
	}

	broken := false
	inputs(func(v any) {
		if broken {
			return
		}
		line, ok := v.(string)
		if !ok {
			broken = true
			err = ErrInputOfAwkMustBeString
			return
		}
		args := []any{line}
		for _, field := range wordSep.Split(strings.Trim(line, " \t"), -1) {
			args = append(args, field)
		}

		newFm := fm.Fork()
		// TODO: Close port 0 of newFm.
		ex := f.Call(newFm, args, eval.NoOpts)
		newFm.Close()

		if ex != nil {
			switch eval.Reason(ex) {
			case nil, eval.Continue:
				// nop
			case eval.Break:
				broken = true
			default:
				broken = true
				err = ex
			}
		}
	})
	return err
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
