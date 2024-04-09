<!-- toc number-sections -->

# Introduction

The Elvish command, `elvish`, contains the Elvish shell and is the main way for
using the Elvish programming language. This documentation describes its behavior
that is not part of the [language](language.html) or any of the standard
modules.

# Using Elvish interactively

Invoking Elvish with no argument runs it in **interactive mode** (unless there
are flags that suppress this behavior). (To use Elvish as your default shell,
see [this page](../get/default-shell.html)).

In this mode, Elvish runs a REPL
([read-eval-print loop](https://en.wikipedia.org/wiki/Read–eval–print_loop))
that evaluates input continuously. The "read" part of the REPL is a rich
interactive editor, and its API is exposed by the [`edit:` module](edit.html).
Each unit of code read is executed as a [code chunk](language.html#code-chunk).

## RC file

Before the REPL starts, Elvish will execute the **RC file**. Its path is
determined as follows:

1.  If the legacy `~/.elvish/rc.elv` exists, it is used (this will be ignored
    starting from 0.21.0).

2.  If the `XDG_CONFIG_HOME` environment variable is defined and non-empty,
    `$XDG_CONFIG_HOME/elvish/rc.elv` is used.

3.  Otherwise, `~/.config/elvish/rc.elv` (non-Windows OSes) or
    `%AppData%\elvish\rc.elv` (Windows) is used.

If the RC file doesn't exist, Elvish does not execute any RC file.

## Database file

Elvish in interactive mode uses a database file to keep command and directory
history. Its path is determined as follows:

1.  If the legacy `~/.elvish/db` exists, it is used (this will be ignored
    starting from 0.21.0).

2.  If the `XDG_STATE_HOME` environment variable is defined and non-empty,
    `$XDG_STATE_HOME/elvish/db.bolt` is used.

3.  Otherwise, `~/.local/state/elvish/db.bolt` (non-Windows OSes) or
    `%LocalAppData%\elvish\db.bolt` is used.

# Running a script

Invoking Elvish with one or more arguments will cause Elvish to execute a script
(unless there are flags that suppress this behavior).

If the `-c` flag is given, the first argument is executed as a single
[code chunk](language.html#code-chunk).

If the `-c` flag is not given, the first argument is taken as a filename, and
the content of the file is executed as a single code chunk.

The remaining arguments are put in [`$args`](builtin.html#$args).

When running a script, Elvish does not evaluate the [RC file](#rc-file).

# Module search directories

When importing [modules](language.html#modules), Elvish searches the following
directories:

1.  If the `XDG_CONFIG_HOME` environment variable is defined and non-empty,
    `$XDG_CONFIG_HOME/elvish/lib` is searched.

    Otherwise, `~/.config/elvish/lib` (non-Window OSes) or
    `%RoamingAppData%\elvish\lib` (Windows) is searched.

2.  If the `XDG_DATA_HOME` environment variable is defined and non-empty,
    `$XDG_DATA_HOME/elvish/lib` is searched.

    Otherwise, `~/.local/share/elvish/lib` (non-Windows OSes) or
    `%LocalAppData%\elvish\lib` (Windows) is searched.

3.  If the `XDG_DATA_DIRS` environment variable is defined and non-empty, it is
    treated as a colon-delimited list of paths (semicolon-delimited on Windows),
    which are all searched.

    Otherwise, `/usr/local/share/elvish/lib` and `/usr/share/elvish/lib` are
    searched on non-Windows OSes. On Windows, no directories are searched.

4.  If the legacy `~/.elvish/lib` directory exists, it is also searched (this
    will be ignored starting from 0.21.0).

# Command-line flags

-   `-buildinfo`: Output information about the Elvish build and quit. See also
    `-version` and `-json`.

-   `-c`: Treat the first argument as code to execute, instead of name of file
    to execute. See [running a script](#running-a-script).

-   `-compileonly`: Parse and compile Elvish code without executing it. Useful
    for checking parse and compilation errors.

    Currently ignored when Elvish is run
    [interactively](#using-elvish-interactively) (so can't be used to check the
    [RC file](#rc-file), for example).

-   `-deprecation-level n`: Show warnings for features deprecated as of version
    0.*n*.

    In release builds, the default value matches the release version, and this
    flag is mainly useful for hiding newly introduced deprecation warnings. For
    example, if you have upgraded from 0.41 to 0.42, you can use
    `-deprecation-level 41` to hide deprecation warnings introduced in 0.42,
    before you have time to fix those warnings.

    In HEAD builds, the default value matches the *previous* release version,
    and this flag is mainly useful for previewing upcoming deprecations. For
    example, if you are running a HEAD version between the 0.42.0 release and
    0.43.0 release, you can use `-deprecation-level 43` to preview deprecations
    that will be introduced in 0.43.0.

-   `-help`: Show usage help and quit.

-   `-i`: A no-op flag, introduced for POSIX compatibility. In future, this may
    be used to force interactive mode.

-   `-json`: Show the output from `-buildinfo`, `-compileonly`, or `-version` in
    JSON.

-   `-log /path/to/log-file`: Path to a file to write debug logs to.

-   `-lsp`: Run the builtin language server.

-   `-norc`: Don't read the [RC file](#rc-file) when running
    [interactively](#using-elvish-interactively). The `-rc` flag is ignored if
    specified.

-   `-rc /path/to/rc`: Path to the [RC file](#rc-file) when running
    [interactively](#using-elvish-interactively). This can be useful for testing
    a new interactive configuration before installing it as your default config.

-   `-version`: Output the Elvish version and quit. See also `-buildinfo` and
    `-json`.

## Daemon flags

The following flags are used by the storage daemon, a process for managing the
access to the [database](#database-file). You shouldn't need to use these flags
unless you are debugging daemon functionalities.

-   `-daemon`: Run the storage daemon instead of an Elvish shell.

-   `-db /path/to/db`: Path to the database file. This only has effect when used
    together with `-daemon`, or when there is no existing daemon running.

-   `-sock /path/to/socket`: Path to the daemon's UNIX socket. A non-daemon
    process will use this socket to send requests to the daemon, while a daemon
    process will listen on this socket.
