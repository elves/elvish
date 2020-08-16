This is the draft release notes for 0.15.0, scheduled to be released on
2021-01-01.

# Deprecated features

The following deprecated features trigger a warning whenever the code is parsed
or compiled, even if it is not executed:

-   The `chr` command is now deprecated. Use `str:from-codepoints` instead.

-   The `ord` command is now deprecated. Use `str:to-codepoints` instead.

-   The `has-prefix` command is now deprecated. Use `str:has-prefix` instead.

-   The `has-suffix` command is now deprecated. Use `str:has-suffix` instead.

The following deprecated features trigger a warning when the code is evaluated:

-   Using `:` in slice indicies is deprecated. Use `..` instead.

# Notable new features

New features in the language:

-   Slice indicies can now use `..` for left-closed, right-open ranges, and
    `..=` for closed ranges.

New features in the standard library:

-   A new `eval` command supports evaluating a dynamic piece of code in a
    restricted namespace.
