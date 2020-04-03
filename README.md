# Elvish: Friendly Interactive Shell and Expressive Programming Language

[![Test Status on Linux](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=Linux&task=Test%20on%20Linux)](https://cirrus-ci.com/github/elves/elvish)
[![Test Status on macOS](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=macOS&task=Test%20on%20macOS)](https://cirrus-ci.com/github/elves/elvish)
[![Test Status on FreeBSD](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=FreeBSD&task=Test%20on%20FreeBSD)](https://cirrus-ci.com/github/elves/elvish)
[![Test status on Windows](https://img.shields.io/appveyor/ci/xiaq/elvish.svg?logo=AppVeyor&label=Windows)](https://ci.appveyor.com/project/xiaq/elvish)

[![Code Coverage on codecov.io](https://img.shields.io/codecov/c/github/elves/elvish.svg?label=codecov)](https://codecov.io/gh/elves/elvish)
[![Code Coverage on coveralls.io](https://img.shields.io/coveralls/github/elves/elvish.svg?label=coveralls)](https://coveralls.io/github/elves/elvish)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/github.com/elves/elvish)
[![GoDoc](https://img.shields.io/badge/godoc-api-blue.svg)](https://godoc.elv.sh)
[![License](https://img.shields.io/badge/BSD-2--clause-blue.svg)](https://github.com/elves/elvish/blob/master/LICENSE)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)

Elvish is a friendly interactive shell and an expressive programming language.
It runs on Linux, BSDs, macOS and Windows. Despite its pre-1.0 status, it is
already suitable for most daily interactive use.

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
