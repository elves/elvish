This is the draft release notes for 0.15.0, scheduled to be released on
2021-01-01.

# Breaking changes

-   Builtin functions and subnamespaces of `edit:` are now read-only.

-   Introspection for rest arguments has changed:

    -   The rest argument is now contained in the `arg-names` field of a
        closure.

    -   The `rest-arg` field now contains the index of the rest argument,
        instead of the name.

-   The `-source` command now runs in a temporary namespace and can no longer
    affect the local scope of its caller.

-   Key modifiers are no longer case insensitive. For example, `Alt` is still
    recognized but `alt` is not. This makes key modifier parsing consistent with
    key names. See [#1163](https://b.elv.sh/1163).

# Deprecated features

The following deprecated features trigger a warning whenever the code is parsed
or compiled, even if it is not executed:

-   The `chr` command is now deprecated. Use `str:from-codepoints` instead.

-   The `ord` command is now deprecated. Use `str:to-codepoints` instead.

-   The `has-prefix` command is now deprecated. Use `str:has-prefix` instead.

-   The `has-suffix` command is now deprecated. Use `str:has-suffix` instead.

-   The `-source` command is now deprecated. Use `eval` instead.

-   The undocumented `esleep` command is now deprecated. Use `sleep` instead.

The following deprecated features trigger a warning when the code is evaluated:

-   Using `:` in slice indices is deprecated. Use `..` instead.

# Notable new features

New features in the language:

-   Slice indices can now use `..` for left-closed, right-open ranges, and `..=`
    for closed ranges.

-   Rest variables and rest arguments are no longer restricted to the last
    variable.

-   Variables containing any character can now be assigned and used by quoting
    their name, for example `'name!' = foo; put $'name!'`.

New features in the standard library:

-   A new `eval` command supports evaluating a dynamic piece of code in a
    restricted namespace.

-   A new `sleep` command.

New features in the interactive editor:

-   SGR escape sequences written from the prompt callback are now supported.

New features in the main program:

-   When using `-compileonly` to check Elvish sources that contain parse errors,
    Elvish will still try to compile the source code and print out compilation
    errors.

# Notable bugfixes

-   Using large lists that contain `$nil` no longer crashes Elvish.
