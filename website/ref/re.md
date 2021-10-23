<!-- toc -->

@module re

# Introduction

The `re:` module wraps Go's `regexp` package. See the Go's doc for supported
[regular expression syntax](https://godoc.org/regexp/syntax).

Function usages notations follow the same convention as the
[builtin module doc](builtin.html).

The following options are supported by multiple functions in this module:

-   `&posix=$false`: Use POSIX ERE syntax. See also
    [doc](http://godoc.org/regexp#CompilePOSIX) in Go package.

-   `&longest=$false`: Prefer leftmost-longest match. See also
    [doc](http://godoc.org/regexp#Regexp.Longest) in Go package.

-   `&max=-1`: If non-negative, limits the maximum number of results.
