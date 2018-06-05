<!-- toc -->

# Introduction

The `re:` module wraps Go's `regexp` package. See the
[doc](http://godoc.org/regexp) for the Go package for syntax of regular
expressions and replacement patterns.

# Functions

Function usages notations follow the same convention as the [builtin module
doc](/ref/builtin.html).

The following options are supported by multiple functions in this module:

*   `&posix` (defaults to `$false`): Use POSIX ERE syntax. See also
    [doc](http://godoc.org/regexp#CompilePOSIX) in Go package.

*   `&longest` (defaults to `$false`): Prefer leftmost-longest match. See also
    [doc](http://godoc.org/regexp#Regexp.Longest) in Go package.

*   `&max` (defaults to -1): If non-negative, maximum number of results.

## find

```elvish
re:find &posix=$false &longest=$false &max=-1 $pattern $source
```

Find all matches of `$pattern` in `$source`.

Each match is represented by a map-like value `$m`; `$m[text]`, `$m[start]`
and `$m[end]` are the text, start and end positions (as byte indicies into
`$source`) of the match; `$m[groups]` is a list of submatches for capture
groups in the pattern. A submatch has a similar structure to a match, except
that it does not have a `group` key. The entire pattern is an implicit capture
group, and it always appears first.

Examples:

```elvish-transcript
~> re:find . ab
▶ [&text=a &start=0 &end=1 &groups=[[&text=a &start=0 &end=1]]]
▶ [&text=b &start=1 &end=2 &groups=[[&text=b &start=1 &end=2]]]
~> re:find '[A-Z]([0-9])' 'A1 B2'
▶ [&text=A1 &start=0 &end=2 &groups=[[&text=A1 &start=0 &end=2] [&text=1 &start=1 &end=2]]]
▶ [&text=B2 &start=3 &end=5 &groups=[[&text=B2 &start=3 &end=5] [&text=2 &start=4 &end=5]]]
```

## match

```elvish
re:match &posix=$false $pattern $source
```

Determine whether `$pattern` matches `$source`. The pattern is not anchored.
Examples:

```elvish-transcript
~> re:match . xyz
▶ $true
~> re:match . ''
▶ $false
~> re:match '[a-z]' A
▶ $false
```

## replace

```elvish
re:replace &posix=$false &longest=$false &literal=$false $pattern $repl $source
```

Replace all occurrences of `$pattern` in `$source` with `$repl`.

The replacement `$repl` can be either

1.  A string-typed replacement template. The template can use `$name` or
    `${name}` patterns to refer to capture groups, where `name` consists of
    letters, digits and underscores. Numbered patterns like `$1` refer to
    capture groups by their order, while named patterns like `$stem` refer to
    capture groups by their names (specified using the syntax
    `(?P<name>...)`). Use `$$` for a literal dollar sign. The name is taken as
    long as possible; for instance, `$1a` is the same as `${1a}`.

    See also doc of Go's regexp package on the [template
    syntax](https://godoc.org/regexp#Regexp.Expand).

2.  A function that takes a string argument and outputs a string. For each
    match, the function is called with the content of the match, and its
    output is used as the replacement.

If `$literal` is true, `$repl` must be a string and is treated literally
instead of as a pattern.

Example:

```elvish-transcript
~> re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'
▶ 'baSH and zSH'
~> re:replace '(ba|z)sh' elvish 'bash and zsh rock'
▶ 'elvish and elvish rock'
~> re:replace '(ba|z)sh' [x]{ put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'
▶ 'BaSh and ZsH'
```

## split

```elvish
re:split &posix=$false &longest=$false &max=-1 $pattern $source
```

Split `$source`, using `$pattern` as separators. Examples:

```elvish-transcript
~> re:split : /usr/sbin:/usr/bin:/bin
▶ /usr/sbin
▶ /usr/bin
▶ /bin
~> re:split &max=2 : /usr/sbin:/usr/bin:/bin
▶ /usr/sbin
▶ /usr/bin:/bin
```

## quote

```elvish
re:quote $string
```

Quote `$string`. Examples:

```elvish-transcript
~> re:quote a.txt
▶ a\.txt
~> re:quote '(*)'
▶ '\(\*\)'
```
