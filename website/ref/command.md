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

Running `elvish -help` lists all supported command-line flags, which are not
repeated here.
