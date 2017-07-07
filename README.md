# A friendly and expressive Unix shell

[![GoDoc](http://godoc.org/github.com/elves/elvish?status.svg)](http://godoc.org/github.com/elves/elvish)
[![Build Status on Travis](https://travis-ci.org/elves/elvish.svg?branch=master)](https://travis-ci.org/elves/elvish)
[![Coverage Status](https://coveralls.io/repos/github/elves/elvish/badge.svg?branch=master)](https://coveralls.io/github/elves/elvish?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/github.com/elves/elvish)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/elves/elvish-public)
[![Telegram Group](https://img.shields.io/badge/telegram%20group-join-blue.svg)](https://telegram.me/elvish)
![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish-000000.svg)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)

Elvish is a cross-platform shell suitable for both interactive use and scripting. It features a full-fledged, non-POSIX-shell programming language with advanced features like namespacing and anonymous functions, and a powerful, fully programmable user interface that works well out of the box.

... which is not 100% true yet. Elvish is already suitable for most daily interactive use, but it is not yet complete. Contributions are more than welcome!

Oh and here is a logo, which happens to be how Elvish looks like when you type `elvish` into it:

![logo](https://elvish.io/assets/logo.svg)

This README documents the development aspect of Elvish. Other information is to be found on the [website](https://elvish.io).


## Building Elvish

To build Elvish, you need

*   A Go toolchain >= 1.6.

*   Linux (with x86, x64 or amd64 CPU) or macOS (with reasonably new hardware).

    It's quite likely that Elvish works on BSDs and other POSIX operating systems, or other CPU architectures; this is not guaranteed due to the lack of good CI support and developers who use such OSes. Pull requests are welcome.

    Windows is **not** supported yet.

### The Correct Way

Elvish is a go-gettable package. To build Elvish, first set up your Go workspace according to [How To Write Go Code](http://golang.org/doc/code.html), and then run

```sh
go get github.com/elves/elvish
```

### The Lazy Way

Here is something you can copy-paste into your terminal:

```sh
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
mkdir -p $GOPATH

go get github.com/elves/elvish

for f in ~/.bashrc ~/.zshrc; do
    printf 'export %s=%s\n' GOPATH '$HOME/go' PATH '$PATH:$GOPATH/bin' >> $f
done
```

The scripts sets up the Go workspace and runs `go get` for you. It assumes that you have a working Go installation and currently use `bash` or `zsh`.

### The Homebrew Way

Users of macOS can build Elvish using [Homebrew](http://brew.sh):

```sh
brew install --HEAD elvish
```


## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the elves have a reputation of high quality. These are usually called *elven* items, but I chose "elvish" because it ends with "sh", a long tradition of Unix shells. It also rhymes with [fish](https://fishshell.com), one of shells that influenced the philosophy of Elvish.

The word "Elvish" should be capitalized like a proper noun. However, when referring to the `elvish` command, use it in lower case with fixed-width font.

Whoever practices the Elvish way by either contributing to it or simply using it is called an **Elf**. (You might have guessed this from the name of the GitHub organization.) The official adjective for Elvish (as in "Pythonic" for Python, "Rubyesque" for Ruby) is **Elven**.
