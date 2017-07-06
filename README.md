# A friendly and expressive Unix shell

[![GoDoc](http://godoc.org/github.com/elves/elvish?status.svg)](http://godoc.org/github.com/elves/elvish)
[![Build Status on Travis](https://travis-ci.org/elves/elvish.svg?branch=master)](https://travis-ci.org/elves/elvish)
[![Coverage Status](https://coveralls.io/repos/github/elves/elvish/badge.svg?branch=master)](https://coveralls.io/github/elves/elvish?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/elves/elvish)](https://goreportcard.com/report/github.com/elves/elvish)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/elves/elvish-public)
[![Telegram Group](https://img.shields.io/badge/telegram%20group-join-blue.svg)](https://telegram.me/elvish)
[![IRC Channel](https://img.shields.io/badge/irc%20channel-join-000000.svg)](irc://irc.freenode.net/elvish)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)

This project aims to explore the potentials of the Unix shell. It is a work in
progress; things will change without warning. The [issues list](https://github.com/elves/elvish/issues) contains many things I'm working on.

This README documents the development aspect of Elvish. All other information is to be found on the [official website](https://elvish.io).

Here is a logo, which happens to be how Elvish looks like when you type `elvish` into it:

![logo](https://elvish.io/assets/logo.svg)


## Building Elvish

Go >= 1.6 is required. Linux is fully supported. It is likely to work on BSDs and Mac OS X. Windows is **not** supported yet.

Elvish is a go-gettable package, and can be installed using `go get github.com/elves/elvish`.

If you are lazy and use `bash` or `zsh` now, here is something you can copy-paste into your terminal:

```sh
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
mkdir -p $GOPATH

go get github.com/elves/elvish

for f in ~/.bashrc ~/.zshrc; do
    printf 'export %s=%s\n' GOPATH '$HOME/go' PATH '$PATH:$GOPATH/bin' >> $f
done
```

[How To Write Go Code](http://golang.org/doc/code.html) explains how `$GOPATH` works.

Users of macOS can build Elvish via [homebrew](http://brew.sh):

```sh
brew install --HEAD elvish
```


## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the elves have a reputation of high quality.  These are usually called *elven* items, but I chose "elvish" because it ends with "sh". It also rhymes with [fish](https://fishshell.com), one of shells that influenced the philosophy of Elvish.

The word "Elvish" should be capitalized like a proper noun. However, when referring to the `elvish` command, use it in lower case with fixed-width font.

Whoever practices the elvish way by either contributing to it or simply using it is called an **elf**. (You might have guessed this from the name of the GitHub organization.) The official adjective for elvish (as in "Pythonic" for Python, "Rubyesque" for Ruby) is "elven".
