# A novel Unix shell

[![GoDoc](http://godoc.org/github.com/elves/elvish?status.svg)](http://godoc.org/github.com/elves/elvish)
[![Build Status](https://travis-ci.org/elves/elvish.svg?branch=master)](https://travis-ci.org/elves/elvish)

This project aims to explore the potentials of the Unix shell. It is a work in
progress; things will change without warning.


## The Interface

Syntax highlighting (also showcasing right-hand-side prompt):

![syntax
highlighting](https://raw.githubusercontent.com/elves/images/master/syntax.png)

Tab completion for files:

![tab completion](https://raw.githubusercontent.com/elves/images/master/completion.png)

Navigation mode (triggered with ^N, inspired by
[ranger](http://ranger.nongnu.org/)):

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
  ~> * (+ 1 2) 3
  ▶ 9
  ```

* Quoting:
  ```
  ~> echo "|  C'est pas une pipe."
  |  C'est pas une pipe.
  ```

* Lists and maps:
  ```
  ~> println list: [a list] map: [&key value]
  list: [a list] map: [&key value]
  ~> println [a b c][0]
  a
  ~> println [&key value][key]
  value
  ```

* Variables:
  ```
  ~> v=[&foo bar]; put $v[foo]
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
  ~> map [x]{/ $x 2} (map [x]{+ 10 $x} [1 2 3])
  [5.5 6 6.5]
  ```

* More natural concatenative style:
  ```
  ~> put 1 2 3 | each [x]{+ 10 $x} | each [x]{/ $x 2}
  ▶ 5.5
  ▶ 6
  ▶ 6.5
  ```

* A separate `env:` namespace for environmental variables:
  ```
  ~> put $env:HOME
  ▶ /home/xiaq
  ~> $env:PATH = $env:PATH":/bin"
  ```

The language is not yet complete. Notably, control structures like `if` and
`while` are not yet implemented. The issues list contain some of things I'm
currently working on.

## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the
elves have a reputation of high quality.  These are usually called **elven**
items, but I chose **elvish** for an obvious reason.

The adjective for elvish is also "elvish", not "elvishy" and definitely not
"elvishish".

It is not directly related to the fictional
[elvish language](https://en.wikipedia.org/wiki/Elvish_language), but I
believe there is not much room for confusion and the google-ability is still
pretty good.


## Building

Go >= 1.4 is required. This repository is a go-getable package.

Linux and FreeBDS are fully supported. It is likely to work on other BSDs and
Mac OS X. Windows is **not** supported yet.

In case you are new to Go, you are advised to read [How To Write Go
Code](http://golang.org/doc/code.html), but here is a quick snippet:

```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
go get github.com/elves/elvish
elvish
```

To update and rebuild:

```
go get -u github.com/elves/elvish
```

Remember to put the two `export`s above into your `bashrc` or `zshrc` (or
whatever).

## Notes for Contributors

### Testing

Always run unit tests before committing. `make` will take care of this.

### Generated files

Some files are generated from other files. They should be commmited into the
repository for this package to be go-getable. Run `go generate ./...` to
regenerate them in case you modified the source.

### Formatting the Code

Always format the code with `goimports` before committing. Run
`go get golang.org/x/tools/cmd/goimports` to install `goimports`, and
`goimports -w .` to format all golang sources.

To automate this you can set up a `goimports` filter for Git by putting this
in `~/.gitconfig`:

    [filter "goimports"]
        clean = goimports
        smudge = cat

Git will then always run `goimports` for you before comitting, since
`.gitattributes` in this repository refers to this filter. More about Git
attributes and filters
[here](https://www.kernel.org/pub/software/scm/git/docs/gitattributes.html).

### Licensing

By contributing, you agree to license your code under the same license as
existing source code of elvish. See the LICENSE file.
