<!-- toc -->

# Introduction

The builtin module contains facilities that are potentially useful to all users.
It occupies the `builtin:` namespace. You rarely have to explicitly specify the
namespace though, since it is one of the namespaces consulted when resolving
unqualified names.

## Usage Notation

The usage of a builtin command is described by giving an example usage, using
variables as arguments. For instance, The `repeat` command takes two arguments
and are described as:

```elvish
repeat $n $v
```

Optional arguments are represented with a trailing `?`, while variadic arguments
with a trailing `...`. For instance, the `count` command takes an optional list:

```elvish
count $input-list?
```

While the `put` command takes an arbitrary number of arguments:

```elvish
put $values...
```

Options are given along with their default values. For instance, the `echo`
command takes an `sep` option and arbitrary arguments:

```elvish
echo &sep=' ' $value...
```

(When you calling functions, options are always optional.)

## Supplying Input

Some builtin functions, e.g. `count` and `each`, can take their input in one of
two ways:

1. From pipe:

    ```elvish-transcript
    ~> put lorem ipsum | count # count number of inputs
    2
    ~> put 10 100 | each [x]{ + 1 $x } # apply function to each input
    ▶ 11
    ▶ 101
    ```

    Byte pipes are also possible; one line becomes one input:

    ```elvish-transcript
    ~> echo "a\nb\nc" | count # count number of lines
    ▶ 3
    ```

1. From an argument -- an iterable value:

    ```elvish-transcript
    ~> count [lorem ipsum] # count number of elements in argument
    2
    ~> each [x]{ + 1 $x } [10 100] # apply to each element in argument
    ▶ 11
    ▶ 101
    ```

    Strings, and in future, other sequence types are also possible:

    ```elvish-transcript
    ~> count lorem
    ▶ 5
    ```

When documenting such commands, the optional argument is always written as
`$input-list?`. On the other hand, a trailing `$input-list?` always indicates
that a command can take its input in one of two ways above: this fact is not
repeated below.

**Note**: You should prefer the first form, unless using it requires explicit
`put` commands. Avoid `count [(some-command)]` or
`each $some-func [(some-command)]`; they are, most of the time, equivalent to
`some-command | count` or `some-command | each $some-func`.

**Rationale**: An alternative way to design this is to make (say) `count` take
an arbitrary number of arguments, and count its arguments; when there is 0
argument, count inputs. However, this leads to problems in code like `count *`;
the intention is clearly to count the number of files in the current directory,
but when the current directory is empty, `count` will wait for inputs. Hence it
is required to put the input in a list: `count [*]` unambiguously supplies input
in the argument, even if there is no file.

## Commands That Operate On Numbers

Commands that operate on numbers are quite flexible about the
format of those numbers. See the discussion of the [number data
type](./language.html#number).

Because numbers are normally specified as strings, rather than as
an explicit `float64` data type, some builtin commands have variants
intended to operate on strings or numbers exclusively. For instance, the
numerical equality command is `==`, while the string equality command is
`==s`. Another example is the `+` builtin, which only operates on numbers
and does not function as a string concatenation command. Consider these
examples:

```elvish-transcript
~/projects/3rd-party/elvish> + x 1
Exception: wrong type of 1'th argument: cannot parse as number: x
[tty], line 1: + x 1
~> + inf 1
▶ (float64 +Inf)
~> + -inf 1
▶ (float64 -Inf)
~> + -infinity 1
▶ (float64 -Inf)
~> + -infinityx 1
Exception: wrong type of 1'th argument: cannot parse as number: -infinityx
[tty], line 1: + -infinityx 1
```

## Predicates

Predicates are functions that write exactly one output that is either `$true` or
`$false`. They are described like "Determine ..." or "Test ...". See [`is`](#is)
for one example.

## "Do Not Use" Functions and Variables

The name of some variables and functions have a leading `-`. This is a
convention to say that it is subject to change and should not be depended upon.
They are either only useful for debug purposes, or have known issues in the
interface or implementation, and in the worst case will make Elvish crash.
(Before 1.0, all features are subject to change, but those ones are sure to be
changed.)

Those functions and variables are documented near the end of the respective
sections. Their known problem is also discussed.

@elvdoc -dir ../pkg/eval
