# A novel Unix shell

[![GoDoc](http://godoc.org/github.com/elves/elvish?status.svg)](http://godoc.org/github.com/elves/elvish)
[![Build Status on Travis](https://travis-ci.org/elves/elvish.svg?branch=master)](https://travis-ci.org/elves/elvish)

This project aims to explore the potentials of the Unix shell. It is a work in
progress; things will change without warning.

The [wiki](https://github.com/elves/elvish/wiki) has a random list of things you might want to know. The [issues list](https://github.com/elves/elvish/issues) contains many things I'm working on.

## The Interface

Syntax highlighting (also showcasing right-hand-side prompt):

![syntax highlighting](https://raw.githubusercontent.com/elves/images/master/syntax.png)

Tab completion for files:

![tab completion](https://raw.githubusercontent.com/elves/images/master/completion.png)

Navigation mode (triggered with ^N, inspired by [ranger](http://ranger.nongnu.org/)):

![navigation mode](https://raw.githubusercontent.com/elves/images/master/navigation.png)


Planned features:

* Auto-suggestion (like fish)
* Programmable line editor
* Directory jumping (#27)
* A vi keybinding that makes sense
* History listing (like
  [ptpython](https://github.com/jonathanslenders/ptpython))
* Intuitive multiline editing

## The Language

Some things that the language is already capable of:

* External programs and pipelines: (`~>` is the prompt):
  ```
  ~> vim README.md
  ...
  ~> cat -v /dev/random
  ...
  ~> dmesg | grep -i acpi
  ...
  ```

* Arithmetics using the prefix notation:
  ```
  ~> + 1 2
  ▶ 3
  ~> mul (+ 1 2) 3
  ▶ 9
  ```

* Quoting:
  ```
  ~> echo "|  C'est pas une pipe."
  |  C'est pas une pipe.
  ```

* Lists and maps:
  ```
  ~> println list: [a list] map: [&key=value]
  list: [a list] map: [&key=value]
  ~> println [a b c][0]
  a
  ~> println [&key=value][key]
  value
  ```

* Variables:
  ```
  ~> v=[&foo=bar]; put $v[foo]
  ▶ bar
  ```

* Defining functions:
  ```
  ~> fn map [f xs]{ put [(put $@xs | each $f)] }
  ```

* Lisp-like functional programming:
  ```
  ~> map [x]{+ 10 $x} [1 2 3]
  [11 12 13]
  ~> map [x]{div $x 2} (map [x]{+ 10 $x} [1 2 3])
  [5.5 6 6.5]
  ```

* More natural concatenative style:
  ```
  ~> put 1 2 3 | each [x]{+ 10 $x} | each [x]{div $x 2}
  ▶ 5.5
  ▶ 6
  ▶ 6.5
  ```

* A separate `env:` namespace for environmental variables:
  ```
  ~> put $env:HOME
  ▶ /home/xiaq
  ~> env:PATH=$env:PATH":/bin"
  ```


## Getting elvish

### Prebuilt binaries

Prebuilt binaries are available for 64-bit [Linux](https://dl.elvish.io/elvish-linux.tar.xz) and [Mac OS X](https://dl.elvish.io/elvish-osx.tar.xz). They are always built using the latest commit that builds. Download the archive and use `sudo tar xfJ elvish-*.tar.xz -C /usr/bin` to install.

### Building It Yourself

Go >= 1.5 is required. This repository is a go-getable package.

Linux is fully supported. It is likely to work on BSDs and Mac OS X. Windows is **not** supported yet.

In case you are new to Go, you are advised to read [How To Write Go Code](http://golang.org/doc/code.html), but here is a quick snippet:

```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
go get github.com/elves/elvish{,/elvish-stub}
elvish
```

To update and rebuild:

```
go get -u github.com/elves/elvish{,/elvish-stub}
```

Remember to put the two `export`s above into your `bashrc` or `zshrc` (or whatever).

## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the elves have a reputation of high quality.  These are usually called **elven** items, but I chose **elvish** for an obvious reason.

The adjective for elvish is also "elvish", not "elvishy" and definitely not "elvishish".

## Test coverages:

|Package|Coverage|
|-------|--------|
|edit|[![edit](https://gocover.io/_badge/github.com/elves/elvish/edit/)](https://gocover.io/github.com/elves/elvish/edit/)|
|eval|[![eval](https://gocover.io/_badge/github.com/elves/elvish/eval/)](https://gocover.io/github.com/elves/elvish/eval/)|
|glob|[![glob](https://gocover.io/_badge/github.com/elves/elvish/glob/)](https://gocover.io/github.com/elves/elvish/glob/)|
|parse|[![parse](https://gocover.io/_badge/github.com/elves/elvish/parse/)](https://gocover.io/github.com/elves/elvish/parse/)|
|store|[![store](https://gocover.io/_badge/github.com/elves/elvish/store/)](https://gocover.io/github.com/elves/elvish/store/)|
|sys|[![sys](https://gocover.io/_badge/github.com/elves/elvish/sys/)](https://gocover.io/github.com/elves/elvish/sys/)|
|util|[![util](https://gocover.io/_badge/github.com/elves/elvish/util/)](https://gocover.io/github.com/elves/elvish/util/)|
