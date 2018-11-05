# Elvish: Friendly Interactive Shell and Expressive Programming Language

[![Build Status on Travis](https://img.shields.io/travis/elves/elvish.svg?logo=travis&label=linux%20%26%20macOS)](https://travis-ci.org/elves/elvish)
[![Build status on AppVeyor](https://img.shields.io/appveyor/ci/xiaq/elvish.svg?logo=appveyor&label=windows)](https://ci.appveyor.com/project/xiaq/elvish) <!-- [![Build Status on VSTS](https://img.shields.io/vso/build/xiaq/13c48a6c-b2dc-472e-af6c-169bf448f8e6/1.svg?logo=tfs&label=macOS)](https://xiaq.visualstudio.com/elvish/_build)
[![Code Coverage on codecov.io](https://img.shields.io/codecov/c/github/elves/elvish.svg?label=codecov)](https://codecov.io/gh/elves/elvish) -->
[![Code Coverage on coveralls.io](https://img.shields.io/coveralls/github/elves/elvish.svg?label=coveralls)](https://coveralls.io/github/elves/elvish)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/github.com/elves/elvish)
[![GoDoc](https://img.shields.io/badge/godoc-api-blue.svg)](https://godoc.elv.sh)
[![License](https://img.shields.io/badge/BSD-2--clause-blue.svg)](https://github.com/elves/elvish/blob/master/LICENSE)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)

Elvish is a friendly interactive shell and an expressive programming language.
It runs on Linux, BSDs, macOS and Windows. Despite its pre-1.0 status, it is
already suitable for most daily interactive use.

Most of the resources for Elvish can be found on the [official website](https://elv.sh).

User groups (all connected thanks to [matterbridge](https://github.com/42wim/matterbridge/)):
[![Gitter](https://img.shields.io/badge/gitter-elves/elvish-blue.svg?logo=gitter-white)](https://gitter.im/elves/elvish)
[![Telegram Group](https://img.shields.io/badge/telegram-@elvish-blue.svg)](https://telegram.me/elvish)
[![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish-blue.svg)](https://webchat.freenode.net/?channels=elvish)

## Building Elvish

To build Elvish, you need

*   Linux, {Free,Net,Open}BSD, macOS, or Windows (Windows support is experimental).

*   Go >= 1.10.

If you have not done so, first set up your environment by following [How To Write Go Code](http://golang.org/doc/code.html).

There are two ways to build Elvish. You can build it directly with `go get`:

```sh
go get github.com/elves/elvish
```

However, binaries built in this way lacks some build-time information; for instance, `elvish -version` will show `unknown`. To add such information, use `make`:

```sh
cd `go env GOPATH`/src/github.com/elves/elvish
make get
```

In either cases, the binary is placed in `$GOPATH/bin`. Consider adding it to your `$PATH` if you want to run the Elvish binary you just built by just typing `elvish`.

See [CONTRIBUTING.md](CONTRIBUTING.md) for more notes for contributors.
You can also join one of the developer groups (also connected together by
matterbridge):
[![Gitter for Developers](https://img.shields.io/badge/gitter-elves/elvish--dev-000000.svg?logo=gitter-white)](https://gitter.im/elves/elvish-dev)
[![Telegram Group for Developers](https://img.shields.io/badge/telegram-@elvish__dev-000000.svg)](https://telegram.me/elvish_dev)
[![#elvish-dev on freenode](https://img.shields.io/badge/freenode-%23elvish--dev-000000.svg)](https://webchat.freenode.net/?channels=elvish-dev)
