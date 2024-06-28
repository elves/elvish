# Elvish

An expressive, versatile cross-platform shell implemented in Go

Qi Xiao (xiaq)

2023-03-22 @ London Gophers

***

# Intro

-   Software engineer at Google

    -   I don't speak for my employer, etc. etc.

-   Started Elvish in 2013

***

# Why build a shell

-   I use a shell every day I use a computer

-   Traditional shells aren't good enough

-   Want a shell with

    -   Nice interactive features

    -   Serious programming constructs

***

# Why not build a shell

-   Too hard for users to switch

    -   Hopefully overcomeable

-   "Shells are *supposed* to be primitive and arcane"

    -   Defeatism

***

# Part 1:

# Elvish as a shell and programming language

***

# A better shell

-   Enhanced interactive features

    -   Location mode (Ctrl-L)

    -   Navigation mode (Ctrl-N)

-   This works like a traditional shell (only new feature is `**`):

    ```elvish
    cat **.go | grep -v '^$' | wc -l
    ```

***

# Somewhat new takes

-   Get command output with `()`:

    ```elvish
    wget dl.elv.sh/(go env GOGOS)-(go env GOARCH)/elvish-HEAD
    ```

-   Variables must be declared:

    ```elvish
    var lorem = foo
    echo $lorme # Error!
    ```

-   Variable values are never split:

    ```elvish
    var x = 'foo bar'
    touch $x
    ```

***

# Data structures

-   Lists: `[foo bar]`

-   Maps: `[&foo=bar]`

-   Index them:

    ```elvish
    var l = [foo bar]; echo $l[0]
    var m = [&foo=bar]; echo $m[foo]
    ```

-   Nest them:

    ```elvish
    var l = [[foo] [&[foo]=[bar]]]
    echo $l[1][[foo]]
    ```

***

# Data structures in "shell use cases"

-   Automation with multiple parameters:

    ```elvish
    for host [[&name=foo &port=22] [&name=bar &port=2200]] {
      scp -P $host[port] elvish root@$host[name]:/usr/local/bin/
    }
    ```

-   Abbreviations are maps:

    ```elvish
    set edit:abbr[xx] = '> /dev/null'
    ```

-   Better programming language leads to better shell

***

# Lambdas!

-   Lambdas:

    ```elvish
    var f = {|x| echo 'Hello '$x }
    $f world
    var g = { echo 'Hello' } # Omit empty arg list
    $g
    ```

***

# Lambdas in "shell use cases"

-   Prompts are lambdas:

    -   ```elvish
        set edit:prompt = { put (whoami)@(tilde-abbr $pwd)'> ' }
        ```

    -   No mini-language like `PS1='\u@\w> '`

-   Configure command completion with a map to lambdas:

    ```elvish
    set edit:completion:arg-completer[foo] = {|@x|
      echo lorem
      echo ipsum
    }
    ```

***

# Outputting and piping values

-   Arithmetic:

    ```elvish
    * 7 (+ 1 5)
    ```

-   String processing:

    ```elvish
    use str
    for x [*.jpg] {
      gm convert $x (str:trim-suffix $x .jpg).png
    }
    ```

-   Stream processing:

    ```elvish
    put [x y] [x] | count
    put [x y] [x] | each {|v| put $v[0] }
    ```

***

# Conclusions of part 1

-   Familiar shell

-   With a real programming language

-   Better programming language -> better shell

-   Many features not covered

    -   Environment variables, exceptions, user-defined modules, ...

***

# Part 2:

# Elvish as a Go project

-   Implementation

-   Experience of using Go

-   CI/CD practice

***

# Implementation overview

-   Frontend (parser)

    -   Hand written parser

    -   Interface and recursion

    -   `type Node interface { parse(*parser) }`

-   Backend (interpreter)

    -   Compile the parse tree into an op tree

    -   Interface and recursion

    -   `type effectOp interface { exec(*Frame) Exception }`

-   Terminal UI

    -   Arcane escape codes

***

# Notable Go features in Elvish's runtime

-   Running external commands: `os.StartProcess`

-   Pipelines: goroutines, `sync.WaitGroup`

-   Outputting and piping values: `chan any`

-   Big numbers: `math/big`

-   Go standard library

    -   Elvish's `str:trim-suffix` is just Go's `strings.TrimSuffix`, etc.

    -   Reflection-based binding

***

# Why Go?

-   Reasonably performant

-   Suitable runtime (goroutines, GC)

-   Fast compilation and easy cross-compilation

-   Rust wasn't released yet

***

# Wishlist

-   Nil safety

-   Plugin support on more platforms

-   Faster reflection

***

# Experience over the years

-   Go 1.1 (!) was the latest version when I started Elvish

-   Relatively few changes over the years

-   1.5: vendoring

-   1.11: modules

-   1.13: `-trimpath`

-   1.16: `//go:embed`

-   1.18: Generics, fuzzing

***

# CI

-   GitHub Actions

    -   Tests, go vet, staticcheck, etc.

    -   Uploading test coverages to codecov.io

-   Cirrus CI for more platforms

    -   {Free Net Open}BSD

    -   Linux ARM64

-   Both are free for Elvish's current use cases

***

# Website and prebuilt binaries

-   <https://elv.sh/>

-   Webhook

-   Building the website

    -   A custom CommonMark implementation

    -   A custom static site generator

-   Building the binaries

    -   Reproducible

    -   Verified by both CI environments

-   Two nodes globally, with geo DNS

-   ~ $20 per month (VPS + domain name + DNS)

***

# Learn more

<https://elv.sh>
