# Elvish: Friendly Interactive Shell and Expressive Programming Language

[![Build Status on Travis](https://img.shields.io/travis/elves/elvish.svg?logo=travis&label=linux%20%26%20macOS)](https://travis-ci.org/elves/elvish)
[![Build status on AppVeyor](https://img.shields.io/appveyor/ci/xiaq/elvish.svg?logo=appveyor&label=windows)](https://ci.appveyor.com/project/xiaq/elvish)
[![Code Coverage on codecov.io](https://img.shields.io/codecov/c/github/elves/elvish.svg?label=codecov)](https://codecov.io/gh/elves/elvish)
[![Code Coverage on coveralls.io](https://img.shields.io/coveralls/github/elves/elvish.svg?label=coveralls)](https://coveralls.io/github/elves/elvish)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/github.com/elves/elvish)
[![GoDoc](https://img.shields.io/badge/godoc-api-blue.svg)](https://godoc.elv.sh)
[![License](https://img.shields.io/badge/BSD-2--clause-blue.svg)](https://github.com/elves/elvish/blob/master/LICENSE)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)

Elvish is a friendly interactive shell and an expressive programming language.
It runs on Linux, BSDs, macOS and Windows. Despite its pre-1.0 status, it is
already suitable for most daily interactive use.

Most of the resources for Elvish can be found on the
[official website](https://elv.sh).

User groups (all connected thanks to
[matterbridge](https://github.com/42wim/matterbridge/)):
[![Gitter](https://img.shields.io/badge/gitter-elves/elvish-blue.svg?logo=gitter-white)](https://gitter.im/elves/elvish)
[![Telegram Group](https://img.shields.io/badge/telegram-@elvish-blue.svg)](https://telegram.me/elvish)
[![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish-blue.svg)](https://webchat.freenode.net/?channels=elvish)

## Building Elvish

To build Elvish from source, you need

-   A supported OS: Linux, {Free,Net,Open}BSD, macOS, or Windows.

    **NOTE**: Windows support is experimental, and only Windows 10 is supported.

-   Go >= 1.13.

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
go get -u github.com/elves/elvish
```

This will clone the Git repository to `~/go/src/github.com/elves/elvish`,
updating it if already exists, and install Elvish to `~/go/bin`. However, Elvish
built this way will lack version information, although it is otherwise fully
functional.

Some tips on installation:

-   Remember to add `$HOME/go/bin` to your `PATH` so that you can run `elvish`
    directly.

-   If you want to install Elvish to a different place, follow
    [these steps](https://github.com/golang/go/wiki/SettingGOPATH) to set
    `GOPATH`, and Elvish will be installed to `$GOPATH/bin` instead.

## Contributing to Elvish

See [CONTRIBUTING.md](CONTRIBUTING.md) for more notes for contributors.
