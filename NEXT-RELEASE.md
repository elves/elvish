This is the draft release notes for the next release, 0.14.0. It is planned to
be released on 2020-07-01, 6 months after 0.13.0.

# Breaking changes

-   The `type` field of the value returned by `src` have been removed.

-   The `all` command no longer preserves byte inputs as is; instead it turns
    them into values, one each line. It also accepts an optional list argument,
    consistent with other value-taking commands.

-   Output captures now split trailing carriage returns from each line,
    effectively making `\r\n` also valid line separators.

# Deprecated features

Elvish now has a deprecation mechanism to give advance notice for breaking
changes. Deprecated features trigger warnings, and will be removed in the next
release.

The following deprecated features trigger a warning at compilation time:

-   The `explode` command is now deprecated. Use `all` instead.

-   The `joins`, `replaces` and `splits` commands are now deprecated. Use
    `str:join`, `str:replace` and `str:split` instead.

-   The `-time` command has been promoted to `time`. The `-time` command is now
    a deprecated alias for `time`.

The following deprecated features trigger a warning at evaluation time:

-   The `&display-suffix` option of the `edit:complex-candidate` is now
    deprecated. Use the `&display` option instead.

The following deprecated features, unfortunately, do not trigger any warnings:

-   The `path` field of the value returned by `src` is now deprecated. Use the
    `name` field instead.

# Notable new features

New features in the language:

-   The printing of floating-point numbers has been tweaked to feel much more
    natural ([#811](https://b.elv.sh/811)).

-   Scripts may now use relative `use` to import files outside `~/.elvish/lib`.

-   Dynamic strings may now be used as command as long as they contain slashes
    ([#764](https://b.elv.sh/764)).

-   Elvish now supports CRLF line endings in source files.

-   Comments are now allowed in list and map literals, and other places where
    newlines serve as separators.

New features in the standard library:

-   A new `order` command for sorting values has been introduced.

-   A new `platform:` module has been introduced.

-   A new `unix:` module has been introduced.

-   A new `math:` module has been introduced.

-   A new `exc:` module has been introduced, including support for printing
    exception stacktraces and determining the cause of the exception.

-   The `fail` command now takes an argument of any type. In particular, if the
    argument is an exception, it rethrows the exception.

-   A new `make-map` command creates a map from a sequence of pairs.

-   A new `read-line` command can be used to read a single line from the byte
    input.

-   The `-time` command has been promoted to `time`, and it now accepts an
    `&on-end` callback to specify how to save the duration of the execution.

-   A new `one` command has been added.

-   A new `read-upto` command can now be added to read byte inputs up to a
    delimiter ([#831](https://b.elv.sh/831)).

New features in the interactive editor:

-   When a callback of the interactive editor throws an exception, the exception
    is now saved in a `$edit:exceptions` variable for closer examination.

-   A new alternative abbreviation mechanism, "small word abbreviation", is now
    available and configurable via `$edit:small-word-abbr`.

-   The ratios of the column widths in navigation mode can now be configured
    with `$edit:navigation:width-ratio` ([#464](https://b.elv.sh/464))

-   A new `$edit:add-cmd-filters` variable is now available for controlling
    whether a command is added to the history.

    The default value of this variable filters out commands that start with a
    space.

-   The `edit:complex-candidate` now supports a `&display` option to specify the
    full display text.

Other improvements:

-   Elvish now uses `$XDG_RUNTIME_DIR` to keep runtime files if possible.

-   Elvish now increments the `$SHLVL` environment variable.

# Notable bugfixes

-   Running elvish with stdin not attached to a tty no longer attempts to start
    an interactive shell ([#661](https://b.elv.sh/661)). Depending on the CLI
    arguments it may treat stdin as a script or data.

-   Invalid option names or values passed to builtin functions now correctly
    trigger an exception, instead of being silently ignored.

-   Elvish no longer crashes when redirecting to a high FD
    ([#788](https://b.elv.sh/788)).

-   Indexing access to nonexistent variables now correctly triggers a
    compilation error ([#889](https://b.elv.sh/889)).

-   The interactive REPL no longer highlights complex commands as red
    ([#881](https://b.elv.sh/881)).

-   Glob patterns after `~username` now evaluate correctly
    ([#793](https://b.elv.sh/793)).

-   On Windows, tab completions for directories no longer add superfluous quotes
    backslashes ([#897](https://b.elv.sh/897)).

-   The `edit:move-dot-left-small-word` command has been fixed to actually move
    by a small word instead of a word.

-   A lot of race conditions have been fixed ([#73](https://b.elv.sh),
    [#754](https://b.elv.sh/754)).
