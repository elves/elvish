# Elvish: Expressive Programming Language + Versatile Interactive Shell

[![Test Status on Linux](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=linux&task=Test%20on%20Linux)](https://cirrus-ci.com/github/elves/elvish/master)
[![Test Status on macOS](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=macos&task=Test%20on%20macOS)](https://cirrus-ci.com/github/elves/elvish/master)
[![Test Status on FreeBSD](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=freebsd&task=Test%20on%20FreeBSD)](https://cirrus-ci.com/github/elves/elvish/master)
[![Test status on Windows](https://img.shields.io/appveyor/ci/xiaq/elvish.svg?logo=AppVeyor&label=windows)](https://ci.appveyor.com/project/xiaq/elvish)
[![Test Coverage](https://img.shields.io/codecov/c/github/elves/elvish.svg?logo=Codecov&label=coverage)](https://codecov.io/gh/elves/elvish)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/src.elv.sh)
[![GoDoc](https://img.shields.io/badge/godoc-api-blue.svg)](https://godoc.elv.sh)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/ElvishShell)

Elvish is an expressive programming language and a versatile interactive shell,
combined into one seamless package. It runs on Linux, BSDs, macOS and Windows.

Despite its pre-1.0 status, it is already suitable for most daily interactive
use.

**Visit the official website https://elv.sh for prebuilt binaries, blog posts,
documentation and other resources.**

User groups (all connected with
[matterbridge](https://github.com/42wim/matterbridge/)):
[![Gitter](https://img.shields.io/badge/gitter-elves/elvish-blue.svg?logo=gitter-white)](https://gitter.im/elves/elvish)
[![Telegram Group](https://img.shields.io/badge/telegram-@elvish-blue.svg)](https://telegram.me/elvish)
[![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish-blue.svg)](https://webchat.freenode.net/?channels=elvish)
[![#elvish:matrix.org](https://img.shields.io/badge/matrix-%23elvish:matrix.org-blue.svg)](https://matrix.to/#/#elvish:matrix.org)

## Building Elvish

Most users can just use [prebuilt binaries](https://elv.sh/get/) and do not need
to build from source.

To build Elvish from source, you need

-   A supported OS: Linux, {Free,Net,Open}BSD, macOS, or Windows.

    **NOTE**: Windows support is experimental, and only Windows 10 is supported.

-   Go >= 1.14.

To build Elvish from source, follow these steps:

```sh
# 1. Start from any directory you want to store Elvish's source code
# 2. Clone the Git repository
git clone https://github.com/elves/elvish
# 3. Change into the repository
cd elvish
# 4. Build and install Elvish
make get
```

This will install Elvish to `~/go/bin`.

Alternatively, you can also just use `go get` to install Elvish:

```sh
go get -u src.elv.sh/cmd/elvish
```

This will clone the Git repository to `~/go/src/src.elv.sh`, updating it if
already exists, and install Elvish to `~/go/bin`. However, Elvish built this way
will lack version information, although it is otherwise fully functional.

Some tips on installation:

-   Remember to add `$HOME/go/bin` to your `PATH` so that you can run `elvish`
    directly.

-   If you want to install Elvish to a different place, follow
    [these steps](https://github.com/golang/go/wiki/SettingGOPATH) to set
    `GOPATH`, and Elvish will be installed to `$GOPATH/bin` instead.

## Running (and Debugging) Elvish

Usage: `elvish` \[_options_\] \[_script_\]

TODO: Document the behavior of interactive, versus, non-interactive shells once
https://github.com/elves/elvish/issues/661 is fixed. Specifically, doing
something like `elvish < x.elv` attempts to instantiate an interactive shell
which causes numerous problems since stdin is not a TTY.

See the `-trace` CLI option for how to gain insight into what elvish is doing.
Use `-trace=cmd` if you are looking for the equivalent of `set -o xtrace` or
`set -x` in shells like bash and ksh.

You can send a `SIGUSR1` signal (on UNIX like systems) to a running `elvish`
binary to cause it to write backtraces of each goroutine to stderr. For example,
`kill -USR1 $pid`.

### Command Line Options

-   `-bin`: Path to the `elvish` binary. Do not use.

-   `-buildinfo`: Output information about how `elvish` was built then quit. See
    also `-json` and `-version`.

-   `-c` _string_: Treat the _string_ argument as Elvish code to execute.
    Additional arguments are bundled into the `$argv` variable.

-   `-compileonly`: Compile the Elvish program without risking any side-effects
    from executing the program. This is useful to identify problems with an
    Elvish program that can be detected at compile time. See also `-json`.

-   `-daemon`: Run the daemon instead of a shell. Do not use.

-   `-db` _/path/to/db_: Path to the interactive database. Do not use.

-   `-deprecation-level` _n_: Show warnings for all features deprecated as of
    version 0._n_; e.g., 0.15. The default value depends on the specific
    `elvish` version you are running.

-   `-json`: Output the information from the `-buildinfo` and `-compileonly`
    options in JSON format.

-   `-norc`: Don't read the _\$HOME/.elvish/rc.elv_ script. This is only
    meaningful for interactive shells.

-   `-port` _n_: The TCP/IP port number of the web backend (default:: 3171). Do
    not use.

-   `-sock` _/path/to/daemon_socket_: Path to the UNIX domain socket used by, or
    to communicate with, the daemon. Do not use.

-   `-trace` _options_: Enable tracing `elvish` behavior. Options can be
    separated by commas and/or spaces; e.g., "all,utc" and "all utc" are both
    acceptable. Each trace message belongs to a specific class. By default each
    trace message includes a timestamp that is the fractional seconds since the
    previous message of the same class was output. You can switch to absolute
    timestamps in the local or UTC timezone by including the `local` or `utc`
    option respectively.

    -   `all`: Enable all message classes. See below for the specific clases
        that can be enabled individually.

    -   `file=`_/a/path_: Write the trace output to _/a/path_. The path can be a
        TTY device; i.e., another terminal window. The default is stderr.

    -   `local`: Output trace messages prefaced by an absolute timestamp in the
        local timezone rather than a timestamp relative to the prior message
        with the same class.

    -   `nframes=`_n_: The number of Go backtrace frames to include in trace
        messages that don't specify an explicit number of frames.

    -   `utc`: Output trace messages prefaced by an absolute timestamp in the
        UTC timezone rather than a timestamp relative to the prior message with
        the same class.

    -   Message classes that can be enabled individually:

        -   `adhoc`: Enable ad-hoc messages.

        -   `cmd`: Enable messages about the execution of Elvish statements.
            This is analogous to the `set -x` (or `set -o xtrace`) feature found
            in POSIX shells like bash and ksh.

        -   `daemon`: Enable messages related to interacting with the daemon.

        -   `eval`: Enable messages related to compiling Elvish programs.

        -   `shell`: Enable messages related to the behavior of the shell. At
            the moment this only includes information about signals.

        -   `store`: Enable messages related to interacting with the interactive
            data store.

        -   `terminal`: Enable messages related to interacting with the
            terminal.

-   `-version`: Output the `elvish` version then quit. See also `-buildinfo`.

-   `-web`: Run the web server rather than a shell. Do not use.

## Contributing to Elvish

See [CONTRIBUTING.md](CONTRIBUTING.md) for more notes for contributors.
