# Elvish: Expressive Programming Language + Versatile Interactive Shell

[![CI status](https://github.com/elves/elvish/workflows/CI/badge.svg)](https://github.com/elves/elvish/actions?query=workflow%3ACI)
[![FreeBSD test status](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=freebsd&task=Test%20on%20FreeBSD)](https://cirrus-ci.com/github/elves/elvish/master)
[![gccgo test status](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=gccgo&task=Test%20on%20gccgo)](https://cirrus-ci.com/github/elves/elvish/master)
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

## Contributing to Elvish

See [CONTRIBUTING.md](CONTRIBUTING.md) for more notes for contributors.
