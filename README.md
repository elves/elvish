# Elvish: Friendly and Expressive Shell

[![logo](https://elv.sh/assets/logo.svg)](https://elv.sh/)

Elvish is a cross-platform shell, supporting Linux, BSDs and Windows. It features an expressive programming language, with features like namespacing and anonymous functions, and a fully programmable user interface with friendly defaults. It is suitable for both interactive use and scripting.

... which is not 100% true yet. Elvish is already suitable for most daily interactive use, but it is neither complete nor stablized. Contributions are more than welcome!

This README documents the development aspect of Elvish. Other information is to be found on the [website](https://elv.sh).

[![Build Status on Travis](https://img.shields.io/travis/elves/elvish.svg?logo=travis&label=linux%20%26%20macOS)](https://travis-ci.org/elves/elvish)
[![Build status on AppVeyor](https://img.shields.io/appveyor/ci/xiaq/elvish.svg?logo=appveyor&label=windows)](https://ci.appveyor.com/project/xiaq/elvish)
[![Build Status on VSTS](https://img.shields.io/vso/build/xiaq/13c48a6c-b2dc-472e-af6c-169bf448f8e6/1.svg?logo=tfs&label=macOS)](https://xiaq.visualstudio.com/elvish/_build)
[![Code Coverage on codecov.io](https://img.shields.io/codecov/c/github/elves/elvish.svg?label=codecov)](https://codecov.io/gh/elves/elvish)
[![Code Coverage on coveralls.io](https://img.shields.io/coveralls/github/elves/elvish.svg?label=coveralls)](https://coveralls.io/github/elves/elvish)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/github.com/elves/elvish)
[![GoDoc](https://img.shields.io/badge/godoc-api-blue.svg)](http://godoc.org/github.com/elves/elvish)
[![License](https://img.shields.io/badge/BSD-2--clause-blue.svg)](https://github.com/elves/elvish/blob/master/LICENSE)

[![Gitter](https://img.shields.io/badge/gitter-elvish--public-blue.svg?logo=gitter-white)](https://gitter.im/elves/elvish-public)
[![Telegram Group](https://img.shields.io/badge/telegram-@elvish-blue.svg)](https://telegram.me/elvish)
[![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish-blue.svg)](https://webchat.freenode.net/?channels=elvish)
[![Gitter for Developers](https://img.shields.io/badge/gitter-elvish--dev-000000.svg?logo=gitter-white)](https://gitter.im/elves/elvish-dev)
[![Telegram Group for Developers](https://img.shields.io/badge/telegram-@elvish__dev-000000.svg)](https://telegram.me/elvish_dev)
[![#elvish-dev on freenode](https://img.shields.io/badge/freenode-%23elvish--dev-000000.svg)](https://webchat.freenode.net/?channels=elvish-dev)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)


## Building Elvish

To build Elvish, you need

*   Linux, {Free,Net,Open}BSD, macOS, or Windows (Windows support is experimental).

*   Go >= 1.8.

Once you have a suitable environment, simply build Elvish with `go get`:

```sh
go get github.com/elves/elvish
```

The binary will be placed in `$GOPATH/bin`. If you haven't configured a
`GOPATH`, it defaults to `~/go`. Refer to [How To Write Go
Code](http://golang.org/doc/code.html) on how to set up workspace for Go.


Users of macOS can also build Elvish using [Homebrew](http://brew.sh):

```sh
brew install --HEAD elvish
```


## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).


## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the elves have a reputation of high quality. These are usually called *elven* items, but I chose "elvish" because it ends with "sh", a long tradition of Unix shells. It also rhymes with [fish](https://fishshell.com), one of the shells that influenced the philosophy of Elvish.

The word "Elvish" should be capitalized like a proper noun. However, when referring to the `elvish` command, use it in lower case with fixed-width font.

Whoever practices the Elvish way by either contributing to it or simply using it is called an **Elf**. (You might have guessed this from the name of the GitHub organization.) The official adjective for Elvish (as in "Pythonic" for Python, "Rubyesque" for Ruby) is **Elven**.
