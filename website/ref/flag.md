<!-- toc -->

@module flag

# Introduction

The `flag:` module provides utilities for parsing command-line flags.

This module supports two different conventions of command-line flags. The
[Go convention](#go-convention) is recommended for Elvish scripts (and followed
by the [Elvish command](command.html) itself). The alternative
[getopt convention](#getopt-convention) is also supported, and useful for
writing scripts that wrap existing programs following this convention.

## Go convention

Each flag looks like `-flag=value`.

For **boolean flags**, `-flag` is equivalent to `-flag=true`. For non-boolean
flags, `-flag` treats the next argument as its value; in other words,
`-flag value` is equivalent to `-flag=value`.

Flag parsing stops before any non-flag argument, or after the flag terminator
`--`.

Examples (`-verbose` is a boolean flag and `-port` is a non-boolean flag):

-   In `-port 10 foo -x`, the `port` flag is `10`, and the rest (`foo -x`) are
    non flag arguments.

-   In `-verbose 10 foo -x`, the `verbose` flag is `true`, and the rest
    (`10 foo -x`) are non-flag arguments.

-   In `-port 10 -- -verbose foo`, the `port` flag is `10`, and the part after
    `--` (`-verbose foo`) are non-flag arguments.

Using `--flag` is supported, and equivalent to `-flag`.

**Note**: Chaining of single-letter flags is not supported: `-rf` is one flag
named `rf`, not equivalent to `-r -f`.

## Getopt convention

A flag may have either or both of the following forms:

-   A short form: a single character preceded by `-`, like `-f`;

-   A long form: a string preceded by `--`, like `--flag`.

A flag may take:

-   No argument, like `-f` or `--flag`;

-   A required argument, like `-f value`, `-fvalue`, `--flag=value` or
    `--flag value`;

-   An optional argument, like `-f`, `-fvalue`, `--flag` or `--flag=value`.

A short flag that takes no argument can be followed immediately by another short
flag. For example, if `-r` takes no arguments, `-rf` is equivalent to `-r -f`.
The other short flag may be followed by more short flags (if it takes no
argument), or its argument (if it takes one). Assuming that `-f` and `-v` take
no arguments while `-p` does, here are some examples:

-   `-rfv` is equivalent to `-r -f -v`.

-   `-rfp80` is equivalent to `-r -f -p 80`.

Some aspects of the behavior can be turned on and off as needed:

-   Optionally, flag parsing stops after seeing the flag terminator `--`.

-   Optionally, flag parsing stops before seeing any non-flag argument. Turning
    this off corresponds to the behavior of GNU's `getopt_long`; turning it on
    corresponds to the behavior of BSD's `getopt_long`.

-   Optionally, only long flags are supported, and they may start with `-`.
    Turning this on corresponds to the behavior of `getopt_long_only` and the
    [Go convention](#go-convention).
