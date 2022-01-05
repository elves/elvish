// Package re implements a regular expression module.
package re

import (
	"regexp"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Ns is the namespace for the re: module.
var Ns = eval.BuildNsNamed("re").
	AddGoFns(map[string]interface{}{
		"quote":   regexp.QuoteMeta,
		"match":   match,
		"find":    find,
		"replace": replace,
		"split":   split,
	}).Ns()

//elvdoc:fn quote
//
// ```elvish
// re:quote $string
// ```
//
// Quote `$string` for use in a pattern. Examples:
//
// ```elvish-transcript
// ~> re:quote a.txt
// ▶ a\.txt
// ~> re:quote '(*)'
// ▶ '\(\*\)'
// ```

//elvdoc:fn match
//
// ```elvish
// re:match &posix=$false $pattern $source
// ```
//
// Determine whether `$pattern` matches `$source`. The pattern is not anchored.
// Examples:
//
// ```elvish-transcript
// ~> re:match . xyz
// ▶ $true
// ~> re:match . ''
// ▶ $false
// ~> re:match '[a-z]' A
// ▶ $false
// ```

type matchOpts struct{ Posix bool }

func (*matchOpts) SetDefaultOptions() {}

func match(opts matchOpts, argPattern, source string) (bool, error) {
	pattern, err := makePattern(argPattern, opts.Posix, false)
	if err != nil {
		return false, err
	}
	return pattern.MatchString(source), nil
}

//elvdoc:fn find
//
// ```elvish
// re:find &posix=$false &longest=$false &max=-1 $pattern $source
// ```
//
// Find all matches of `$pattern` in `$source`.
//
// Each match is represented by a map-like value `$m`; `$m[text]`, `$m[start]` and
// `$m[end]` are the text, start and end positions (as byte indices into `$source`)
// of the match; `$m[groups]` is a list of submatches for capture groups in the
// pattern. A submatch has a similar structure to a match, except that it does not
// have a `group` key. The entire pattern is an implicit capture group, and it
// always appears first.
//
// Examples:
//
// ```elvish-transcript
// ~> re:find . ab
// ▶ [&text=a &start=0 &end=1 &groups=[[&text=a &start=0 &end=1]]]
// ▶ [&text=b &start=1 &end=2 &groups=[[&text=b &start=1 &end=2]]]
// ~> re:find '[A-Z]([0-9])' 'A1 B2'
// ▶ [&text=A1 &start=0 &end=2 &groups=[[&text=A1 &start=0 &end=2] [&text=1 &start=1 &end=2]]]
// ▶ [&text=B2 &start=3 &end=5 &groups=[[&text=B2 &start=3 &end=5] [&text=2 &start=4 &end=5]]]
// ```

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

//elvdoc:fn replace
//
// ```elvish
// re:replace &posix=$false &longest=$false &literal=$false $pattern $repl $source
// ```
//
// Replace all occurrences of `$pattern` in `$source` with `$repl`.
//
// The replacement `$repl` can be any of the following:
//
// -   A string-typed replacement template. The template can use `$name` or
//     `${name}` patterns to refer to capture groups, where `name` consists of
//     letters, digits and underscores. A purely numeric patterns like `$1`
//     refers to the capture group with the corresponding index; other names
//     refer to capture groups named with the `(?P<name>...)`) syntax.
//
//     In the `$name` form, the name is taken to be as long as possible; `$1` is
//     equivalent to `${1x}`, not `${1}x`; `$10` is equivalent to `${10}`, not `${1}0`.
//
//     To insert a literal `$`, use `$$`.
//
// -   A function that takes a string argument and outputs a string. For each
//     match, the function is called with the content of the match, and its output
//     is used as the replacement.
//
// If `$literal` is true, `$repl` must be a string and is treated literally instead
// of as a pattern.
//
// Example:
//
// ```elvish-transcript
// ~> re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'
// ▶ 'baSH and zSH'
// ~> re:replace '(ba|z)sh' elvish 'bash and zsh rock'
// ▶ 'elvish and elvish rock'
// ~> re:replace '(ba|z)sh' {|x| put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'
// ▶ 'BaSh and ZsH'
// ```

type replaceOpts struct {
	Posix   bool
	Longest bool
	Literal bool
}

func (*replaceOpts) SetDefaultOptions() {}

func replace(fm *eval.Frame, opts replaceOpts, argPattern string, argRepl interface{}, source string) (string, error) {

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
				return repl.Call(fm, []interface{}{s}, eval.NoOpts)
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

//elvdoc:fn split
//
// ```elvish
// re:split &posix=$false &longest=$false &max=-1 $pattern $source
// ```
//
// Split `$source`, using `$pattern` as separators. Examples:
//
// ```elvish-transcript
// ~> re:split : /usr/sbin:/usr/bin:/bin
// ▶ /usr/sbin
// ▶ /usr/bin
// ▶ /bin
// ~> re:split &max=2 : /usr/sbin:/usr/bin:/bin
// ▶ /usr/sbin
// ▶ /usr/bin:/bin
// ```

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
