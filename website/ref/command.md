<!-- toc number-sections -->

# Introduction

The Elvish command, `elvish`, contains the Elvish shell and is the main way for
using the Elvish programming language. This documentation describes its behavior
that is not part of the [language](language.html) or any of the standard
modules.

# Using Elvish interactively

Invoking Elvish with no argument runs it in **interactive mode** (unless there
are flags that suppress this behavior).

In this mode, Elvish runs a REPL
([read-eval-print loop](https://en.wikipedia.org/wiki/Read–eval–print_loop))
that evaluates input continuously. The "read" part of the REPL is a rich
interactive editor, and its API is exposed by the [`edit:` module](edit.html).
Each unit of code read is executed as a [code chunk](language.html#code-chunk).

## RC file

Before the REPL starts, Elvish will execute the **RC file**. Its path is
determined as follows:

-   If the legacy `~/.elvish/rc.elv` exists, it is used (this will be ignored in
    a future version).

-   Otherwise:

    -   On UNIX (including macOS), `$XDG_CONFIG_HOME/elvish/rc.elv` is used,
        defaulting to `~/.config/elvish/rc.elv` if `$XDG_CONFIG_HOME` is unset
        or empty.

    -   On Windows, `%AppData%\elvish\rc.elv` is used.

If the RC file doesn't exist, Elvish does not execute any RC file.

## Database file

Elvish in interactive mode uses a database file to keep command and directory
history. Its path is determined as follows:

-   If the legacy `~/.elvish/db` exists, it is used (this will be ignored in a
    future version).

-   Otherwise:

    -   On UNIX (including macOS), `$XDG_DATA_HOME/elvish/db.bolt` is used,
        defaulting to `~/.local/state/elvish/db.bolt` if `$XDG_DATA_HOME` is
        unset or empty.

    -   On Windows, `%LocalAppData%\elvish\db.bolt` is used.

# Running a script

Invoking Elvish with one or more arguments will cause Elvish to execute a script
(unless there are flags that suppress this behavior).

If the `-c` flag is given, the first argument is executed as a single
[code chunk](language.html#code-chunk).

If the `-c` flag is not given, the first argument is taken as a filename, and
the content of the file is executed as a single code chunk.

The remaining arguments are put in [`$args`](builtin.html#args).

When running a script, Elvish does not evaluate the [RC file](#rc-file).

# Module search directories

When importing [modules](language.html#modules), Elvish searches the following
directories:

-   On UNIX:

    1. `$XDG_CONFIG_HOME/elvish/lib`, defaulting to `~/.config/elvish/lib` if
       `$XDG_CONFIG_HOME` is unset or empty;

    2. `$XDG_DATA_HOME/elvish/lib`, defaulting to `~/.local/share/elvish/lib` if
       `$XDG_DATA_HOME` is unset or empty;

    3. Paths specified in the colon-delimited `$XDG_DATA_DIRS`, followed by
       `elvish/lib`, defaulting to `/usr/local/share/elvish/lib` and
       `/usr/share/elvish/lib` if `$XDG_DATA_DIRS` is unset or empty.

-   On Windows: `%AppData%\elvish\lib`, followed by `%LocalAppData%\elvish\lib`.

If the legacy `~/.elvish/lib` directory exists, it is also searched.

# Other command-line flags

-   `-buildinfo`: Output information about how `elvish` was built then quit. See
    also `-version` and `-json`. For example:

    ```
    ~> elvish -buildinfo
    Version: 0.18.0-dev.4350ad48b1bb798068ec6a01f149382cd0607e2f
    Go version: go1.17.2
    Reproducible build: true
    ```

-   `-compileonly`: Compile the Elvish program without risking any side-effects
    from executing the program. This is useful to identify problems with an
    Elvish program that can be detected at compile time. At present it cannot be
    used for verifying that the interactive config is syntactically valid
    without executing the config. See also `-json`.

-   `-cpuprofile` _/path/to/profile_: Write a Go CPU profile to the specified
    file. The profile data can be explored using `go tool pprof`. See
    [this](https://go.dev/blog/pprof) blog article for an exploration of how to
    work with Go CPU profiles.

-   `-deprecation-level` _n_: Show warnings for features deprecated as of
    version 0._n_; e.g., 0.18. The default value is the specific `elvish`
    version you are running. Using this flag can be useful when you have
    installed a new version of Elvish but have not yet updated your scripts to
    conform to the new syntax and grammar supported by that new version. In
    practice this is only useful for suppressing warnings related to the most
    recent version by specifying the minor number of the previous version;
    however, that may change in the future.

-   `-help`: Show usage information and quit.

-   `-i`: Force interactive mode. This is silently ignored. It is supported to
    improve compatibility with how POSIX shells are sometimes instantiated.

-   `-json`: Output the information from the `-buildinfo`, `-compileonly`, and
    `-version` flags in JSON format.

-   `-log` _/path/to/log-file_: Write information about the behavior of Elvish
    to the named log file. This does not affect the Elvish daemon which always
    writes to a system specific path.

    TODO: Document the rules for the daemon log file path name.

-   `-norc`: Don't read the [default RC file](#rc-file). This is only meaningful
    for interactive shells.

-   `-rc` _/path/to/rc_: Use _/path/to/rc_ rather than the
    [default RC file](#rc-file). This can be useful for testing a new
    interactive configuration before installing it as your default config.

-   `-version`: Output the `elvish` version then quit. See also `-buildinfo` and
    `-json`.

# Internal use only command-line flags

These flags are internal to Elvish. Do not use them unless you are a developer
who knows what you're doing.

-   `-bin` _/path/to/elvish_: Path to the `elvish` binary.

-   `-daemon`: Run the daemon instead of an Elvish shell.

-   `-db` _/path/to/db_: Path to the interactive database.

-   `-sock` _/path/to/daemon_socket_: Path to the UNIX domain socket used by, or
    to communicate with, the daemon.
