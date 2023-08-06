<!-- toc -->

@module builtin

# Introduction

The builtin module contains facilities that are potentially useful to all users.

## Using builtin: explicitly

The builtin module is consulted implicitly when
[resolving unqualified names](language.html#scoping-rule), and Elvish's
namespacing mechanism makes it impossible for other modules to redefine builtin
symbols. It's almost always sufficient (and safe) to use builtin functions and
variables with their unqualified names.

Nonetheless, the builtin module is also available as a
[pre-defined module](language.html#pre-defined-modules). It can be imported with
`use builtin`, which makes all the builtin symbols available under the
`builtin:` namespace. This can be useful in several cases:

-   To refer to a builtin function when it is shadowed locally. This is
    especially useful when the function that shadows the builtin one is a
    wrapper:

    ```elvish
    use builtin
    fn cd {|@args|
        echo running my cd function
        builtin:cd $@args
    }
    ```

    Note that the shadowing of `cd` is only in effect in the local lexical
    scope.

-   To introspect the builtin module, for example `keys $builtin:`.

## Usage Notation

The usage of a builtin command is described by giving an example usage, using
variables as arguments. For instance, the `repeat` command takes two arguments
and is described as:

```elvish
repeat $n $v
```

Optional arguments are represented with a trailing `?`, while variadic arguments
with a trailing `...`. For instance, the `count` command takes an optional list:

```elvish
count $inputs?
```

While the `put` command takes an arbitrary number of arguments:

```elvish
put $values...
```

Options are given along with their default values. For instance, the `echo`
command takes a `sep` option and arbitrary arguments:

```elvish
echo &sep=' ' $value...
```

(When you call functions, options are always optional.)

## Commands taking value inputs {#value-inputs}

Most commands that take value inputs (e.g. `count`, `each`) can take the inputs
in one of two ways:

1.  From the pipeline:

    ```elvish-transcript
    ~> put lorem ipsum | count # count number of inputs
    2
    ~> put 10 100 | each {|x| + 1 $x } # apply function to each input
    ▶ (num 11)
    ▶ (num 101)
    ```

    If the previous command outputs bytes, one line becomes one string input, as
    if there is an implicit [`from-lines`]() (this behavior is subject to
    change):

    ```elvish-transcript
    ~> print "a\nb\nc\n" | count # count number of lines
    ▶ 3
    ~> use str
    ~> print "a\nb\nc\n" | each $str:to-upper~ # apply to each line
    ▶ A
    ▶ B
    ▶ C
    ```

2.  From an argument -- an iterable value:

    ```elvish-transcript
    ~> count [lorem ipsum] # count number of elements in argument
    2
    ~> each {|x| + 1 $x } [10 100] # apply to each element in argument
    ▶ 11
    ▶ 101
    ```

    Strings, and in future, other sequence types are also supported:

    ```elvish-transcript
    ~> count lorem
    ▶ 5
    ```

When documenting such commands, the optional argument is always written as
`$inputs?`.

**Note**: You should prefer the first form, unless using it requires explicit
`put` commands. Avoid `count [(some-command)]` or
`each $some-func [(some-command)]`; they are equivalent to
`some-command | count` or `some-command | each $some-func`.

**Rationale**: An alternative way to design this is to make (say) `count` take
an arbitrary number of arguments, and count its arguments; when there is 0
argument, count inputs. However, this leads to problems in code like `count *`;
the intention is clearly to count the number of files in the current directory,
but when the current directory is empty, `count` will wait for inputs. Hence it
is required to put the input in a list: `count [*]` unambiguously supplies input
in the argument, even if there is no file.

## Numeric commands

Wherever a command expects a number argument, that argument can be supplied
either with a [typed number](language.html#number) or a string that can be
converted to a number. This includes numeric comparison commands like `==`.

When a command outputs numbers, it always outputs a typed number.

Examples:

```elvish-transcript
~> + 2 10
▶ (num 12)
~> == 2 (num 2)
▶ $true
```

### Exactness-preserving commands {#exactness-preserving}

Some numeric commands are designated **exactness-preserving**. When such
commands are called with only [exact numbers](./language.html#exactness) (i.e.
integers or rationals), they will always output an exact number. Examples:

```elvish-transcript
~> + 10 1/10
▶ (num 101/10)
~> * 12 5/17
▶ (num 60/17)
```

If the condition above is not satisfied -- i.e. when a numeric command is not
designated exactness-preserving, or when at least one of the arguments is
inexact (i.e. a floating-point number), the result is an inexact number, unless
otherwise documented. Examples:

```elvish-transcript
~> + 10 0.1
▶ (num 10.1)
~> + 10 1e1
▶ (num 20.0)
~> use math
~> math:sin 1
▶ (num 0.8414709848078965)
```

There are some cases where the result is exact despite the use of inexact
arguments or non-exactness-preserving commands. Such cases are always documented
in their respective commands.

## Unstable features

The name of some variables and functions have a leading `-`. This is a
convention to say that it is subject to change and should not be depended upon.
They are either only useful for debug purposes, or have known issues in the
interface or implementation, and in the worst case will make Elvish crash.
(Before 1.0, all features are subject to change, but those ones are sure to be
changed.)
