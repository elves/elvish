<!-- toc -->

# Introduction

The builtin module contains facilities that are potentially useful to all users.

## Using builtin: explicitly

The builtin module is consulted implicitly when resolving unqualified names, so
you usually don't need to specify `builtin:` explicitly. However, there are some
cases where it is useful to do that:

-   When a builtin function is shadowed by a local function, you can still use
    the builtin function by specifying `builtin:`. This is especially useful
    when wrapping a builtin function:

    ```elvish
    use builtin
    fn cd [@args]{
        echo running my cd function
        builtin:cd $@args
    }
    ```

-   Introspecting the builtin module, for example `keys $builtin:`.

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

## Numeric commands

Anywhere a command expects a number argument, that argument can be supplied
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

If the condition above is not satisfied - i.e. when a numeric command is not
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
