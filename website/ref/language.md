<!-- toc number-sections -->

# Introduction

This document describes the Elvish programming language. It is both a
specification and an advanced tutorial. The parts of this document marked with
either **notes** or called out as **examples** are non-normative, and only serve
to help you understand the more formal descriptions.

Examples in this document might use constructs that have not yet been
introduced, so some familiarity with the language is assumed. If you are new to
Elvish, start with the [learning materials](../learn/).

# Source code encoding

Elvish source code must be Unicode text encoded in UTF-8.

In this document, **character** is a synonym of
[Unicode codepoint](https://en.wikipedia.org/wiki/Code_point) or its UTF-8
encoding.

# Lexical elements

## Whitespace

In this document, an **inline whitespace** is any of the following:

-   A space (U+0020);

-   A tab (U+0009);

-   A comment: starting with `#` and ending before (but not including) the next
    carriage return, newline or end of file;

-   A line continuation: a `^` followed by a newline (`"\n"`), or a carriage
    return and newline (`"\r\n"`).

A **whitespace** is any of the following:

-   An inline whitespace;

-   A carriage return (U+000D);

-   A newline (U+000A).

## Metacharacters

The following **metacharacters** serve to introduce or delimit syntax
constructs:

| Metacharacter | Use                                                         |
| ------------- | ----------------------------------------------------------- |
| `$`           | Referencing variables                                       |
| `*` and `?`   | Forming wildcard                                            |
| `\|`          | Separating forms in a pipeline                              |
| `&`           | Marking background pipelines; introducing key-value pairs   |
| `;`           | Separating pipelines                                        |
| `<` and `>`   | Introducing IO redirections                                 |
| `(` and `)`   | Enclosing output captures                                   |
| `[` and `]`   | Enclosing list literals, map literals or function signature |
| `{` and `}`   | Enclosing lambda literals or brace expressions              |

The following characters are parsed as metacharacters under certain conditions:

-   `~` is a metacharacter if it appears at the beginning of a compound
    expression, in which case it is subject to
    [tilde expansion](#tilde-expansion);

-   `=` is a metacharacter when used for terminating [map keys](#map) or option
    keys, or denoting [legacy assignment form](#legacy-assignment-form) or
    [temporary assignments](#temporary-assignment).

## Single-quoted string

A single-quoted string consists of zero or more characters enclosed in single
quotes (`'`). All enclosed characters represent themselves, except the single
quote.

Two consecutive single quotes are handled as a special case: they represent one
single quote, instead of terminating a single-quoted string and starting
another.

**Examples**: `'*\'` evaluates to `*\`, and `'it''s'` evaluates to `it's`.

## Double-quoted string

A double-quoted string consists of zero or more characters enclosed in double
quotes (`"`). All enclosed characters represent themselves, except backslashes
(`\`), which introduces **escape sequences**. Double quotes are not allowed
inside double-quoted strings, except after backslashes.

The following escape sequences are supported:

-   `\cX`, where _X_ is a character with codepoint between 0x40 and 0x5F,
    represents the codepoint that is 0x40 lower than _X_. For example, `\cI` is
    the tab character: 0x49 (`I`) - 0x40 = 0x09 (tab). There is one special
    case: A question-mark is converted to del; i.e., `\c?` or `\^?` is
    equivalent to `\x7F`.

-   `\^X` is the same as `\cX`.

-   `\[0..7][0..7][0..7]` is a byte written as an octal value. There must be
    three octal digits following the backslash. For example, `\000` is the nul
    character, and `\101` is the same as `A`, but `\0` is an invalid escape
    sequence (too few digits).

-   `\x..` is a Unicode code point represented by two hexadecimal digits.

-   `\u....` is a Unicode code point represented by four hexadecimal digits.

-   `\U......` is a Unicode code point represented by eight hexadecimal digits.

-   The following single character escape sequences:

    -   `\a` is the "bell" character, equivalent to `\007` or `\x07`.

    -   `\b` is the "backspace" character, equivalent to `\010` or `\x08`.

    -   `\f` is the "form feed" character, equivalent to `\014` or `\x0c`.

    -   `\n` is the "new line" character, equivalent to `\012` or `\x0a`.

    -   `\r` is the "carriage return" character, equivalent to `\015` or `\x0d`.

    -   `\t` is the "tab" character, equivalent to `\011` or `\x09`.

    -   `\v` is the "vertical tab" character, equivalent to `\013` or `\x0b`.

    -   `\\` is the "backslash" character, equivalent to `\134` or `\x5c`.

    -   `\"` is the "double-quote" character, equivalent to `\042` or `\x22`.

An unsupported escape sequence results in a parse error.

**Note**: Unlike most other shells, double-quoted strings in Elvish do not
support interpolation. For instance, `"$name"` simply evaluates to a string
containing `$name`. To get a similar effect, simply concatenate strings: instead
of `"my name is $name"`, write `"my name is "$name`. Under the hood this is a
[compounding](#compounding) operation.

## Bareword

A string can be written without quoting -- a **bareword**, if it only includes
the characters from the following set:

-   ASCII letters (a-z and A-Z) and numbers (0-9);

-   The symbols `!%+,-./:@\_`;

-   Non-ASCII codepoints that are printable, as defined by
    [unicode.IsPrint](https://godoc.org/unicode#IsPrint) in Go's standard
    library.

**Examples**: `a.txt`, `long-bareword`, `elf@elv.sh`, `/usr/local/bin`,
`你好世界`.

Moreover, `~` and `=` are allowed to appear without quoting when they are not
parsed as [metacharacters](#metacharacters).

**Note**: since the backslash (`\`) is a valid bareword character in Elvish, it
cannot be used to escape metacharacter. Use quotes instead: for example, to echo
a star, write `echo "*"` or `echo '*'`, not `echo \*`. The last command just
writes out `\*`.

# Value types

## String

A string is a (possibly empty) sequence of bytes.

[Single-quoted string literals](#single-quoted-string),
[double-quoted string literals](#double-quoted-string)and [barewords](#bareword)
all evaluate to string values. Unless otherwise noted, different syntaxes of
string literals are equivalent in the code. For instance, `xyz`, `'xyz'` and
`"xyz"` are different syntaxes for the same string with content `xyz`.

Strings that contain UTF-8 encoded text can be [indexed](#indexing) with a
**byte index** where a codepoint starts, which results in the codepoint that
starts there. The index can be given as either a typed [number](#number), or a
string that parses to a number. Examples:

-   In the string `elv`, every codepoint is encoded with only one byte, so 0, 1,
    2 are all valid indices:

    ```elvish-transcript
    ~> put elv[0]
    ▶ e
    ~> put elv[1]
    ▶ l
    ~> put elv[2]
    ▶ v
    ```

-   In the string `世界`, each codepoint is encoded with three bytes. The first
    codepoint occupies byte 0 through 2, and the second occupies byte 3 through
    5\. Hence valid indices are 0 and 3:

    ```elvish-transcript
    ~> put 世界[0]
    ▶ 世
    ~> put 世界[3]
    ▶ 界
    ```

Such strings may also be indexed with a slice (see documentation of
[list](#list) for slice syntax). The range determined by the slice is also
interpreted as byte indices, and the range must begin and end at codepoint
boundaries.

The behavior of indexing a string that does not contain valid UTF-8-encoded
Unicode text is unspecified.

**Note**: String indexing will likely change.

## Number

Elvish supports several types of numbers. There is no literal syntax, but they
can be constructed by passing their **string representation** to the
[`num`](builtin.html#num) builtin command:

-   **Integers** are written in decimal (e.g. `10`), hexadecimal (e.g. `0xA`),
    octal (e.g. `0o12`) or binary (e.g. `0b1010`).

-   **Rationals** are written as two exact integers joined by `/`, e.g. `1/2` or
    `0x10/100` (16/100).

-   **Floating-point numbers** are written with a decimal point (e.g. `10.0`) or
    using scientific notation (e.g. `1e1` or `1.0e1`). There are also three
    additional special floating-point values: `+Inf`, `-Inf` and `NaN`.

Digits may be separately by underscores, which are ignored; this permits
separating the digits into groups to improve readability. For example, `1000000`
and `1_000_000` are equivalent, so are `1.234_56e3` and `1.23456e3`.

The string representation is case-insensitive.

### Strings and numbers

Strings and numbers are distinct types; for example, `2` and `(num 2)` are
distinct values.

However, by convention, all language constructs that expect numbers (e.g. list
indices) also accept strings that can be converted to numbers. This means that
most of the time, you can just use the string representation of numbers, instead
of explicitly constructing number values. Builtin
[numeric commands](./builtin.html#numeric-commands) follow the same convention.

When the word **number** appears unqualified in other sections of this document,
it means either an explictly number-typed value (**typed number**), or its
string representation.

When a typed number is converted to a string (e.g. with `to-string`), the result
is guaranteed to convert back to the original number. In other words,
`eq $x (num (to-string $x))` always outputs `$true` if `$x` is a typed number.

### Exactness

Integers and rationals are **exact** numbers; their precision is only limited by
the available memory, and many (but not all) operations on them are guaranteed
to produce mathematically correct results.

Floating-point numbers are IEE 754 double-precision. Since operations on
floating-point numbers in general are not guaranteed to be precise, they are
always considered **inexact**.

For more details, see also the discussion on
[exactness-preserving commands](./builtin.html#exactness-preserving-commands).

## List

A list is a value containing a sequence of values. Values in a list are called
its **elements**. Each element has an index, starting from zero.

List literals are surrounded by square brackets `[ ]`, with elements separated
by whitespace. Examples:

```elvish-transcript
~> put [lorem ipsum]
▶ [lorem ipsum]
~> put [lorem
        ipsum
        foo
        bar]
▶ [lorem ipsum foo bar]
```

**Note**: In Elvish, commas have no special meanings and are valid bareword
characters, so don't use them to separate elements:

```elvish-transcript
~> li = [a, b]
~> put $li
▶ [a, b]
~> put $li[0]
▶ a,
```

A list can be [indexed](#indexing) with the index of an element to obtain the
element, which can take one of two forms:

-   A non-negative integer, an offset counting from the beginning of the list.
    For example, `$li[0]` is the first element of `$li`.

-   A negative integer, an offset counting from the back of the list. For
    instance, `$li[-1]` is the last element `$li`.

In both cases, the index can be given either as a [typed number](#number) or a
number-like string.

A list can also be indexed with a **slice** to obtain a sublist, which can take
one of two forms:

-   A slice `$a..$b`, where both `$a` and `$b` are integers. The result is
    sublist of `$li[$a]` up to, but not including, `$li[$b]`. For instance,
    `$li[4..7]` equals `[$li[4] $li[5] $li[6]]`, while `$li[1..-1]` contains all
    elements from `$li` except the first and last one.

    Both integers may be omitted; `$a` defaults to 0 while `$b` defaults to the
    length of the list. For instance, `$li[..2]` is equivalent to `$li[0..2]`,
    `$li[2..]` is equivalent to `$li[2..(count $li)]`, and `$li[..]` makes a
    copy of `$li`. The last form is rarely useful, as lists are immutable.

    Note that the slice needs to be a **single** string, so there cannot be any
    spaces within the slice. For instance, `$li[2..10]` cannot be written as
    `$li[2.. 10]`; the latter contains two indices and is equivalent to
    `$li[2..] $li[10]` (see [Indexing](#indexing)).

-   A slice `$a..=$b`, which is similar to `$a..$b`, but includes `$li[$b]`.

Examples:

```elvish-transcript
~> li = [lorem ipsum foo bar]
~> put $li[0]
▶ lorem
~> put $li[-1]
▶ bar
~> put $li[0..2]
▶ [lorem ipsum]
```

## Map

A map is a value containing unordered key-value pairs.

Map literals are surrounded by square brackets; a key/value pair is written
`&key=value` (reminiscent to HTTP query parameters), and pairs are separated by
whitespaces. Whitespaces are allowed after `=`, but not before `=`. Examples:

```elvish-transcript
~> put [&foo=bar &lorem=ipsum]
▶ [&foo=bar &lorem=ipsum]
~> put [&a=   10
        &b=   23
        &sum= (+ 10 23)]
▶ [&a=10 &b=23 &sum=33]
```

The literal of an empty map is `[&]`.

Specifying a key without `=` or a value following it is equivalent to specifying
`$true` as the value. Specifying a key with `=` but no value following it is
equivalent to specifying the empty string as the value. Example:

```elvish-transcript
~> echo [&a &b=]
[&a=$true &b='']
```

A map can be indexed by any of its keys. Unlike strings and lists, there is no
support for slices, and `..` and `..=` have no special meanings. Examples:

```elvish-transcript
~> map = [&a=lorem &b=ipsum &a..b=haha]
~> echo $map[a]
lorem
~> echo $map[a..b]
haha
```

You can test if a key is present using [`has-key`](./builtin.html#has-key) and
enumerate the keys using the [`keys`](./builtin.html#keys) builtins.

## Pseudo-map

A pseudo-map is not a single concrete data type. It refers to concrete types
that behave like maps with some restrictions.

A pseudo-map has a fixed set of keys whose values can be accessed by
[indexing](#indexing) like you would for a regular [map](#map). Similarly, you
can use commands like [`keys`](./builtin.html#keys) and
[`has-key`](./builtin.html#keys) on such objects.

Unlike a normal map, it is currently not possible to create a modified version
of an existing pseudo-map: it is not possible to create a pseudo-map with new
keys, without existing keys, or with a different value for a given key.

The pseudo-map mechanism is often used for introspection. For example,
[exceptions](#exception), [user-defined functions](#function), and
[`$buildinfo`](./builtin.html#buildinfo) are pseudo-maps.

## Nil

The value `$nil` serves as the initial value of variables that are declared but
not assigned.

## Boolean

There are two boolean values, `$true` and `$false`.

When converting non-boolean values to the boolean type, `$nil` and exceptions
convert to `$false`. All the other non-boolean values convert to `$true`.

## Exception

An exception carries information about errors during the execution of code.

There is no literal syntax for exceptions. See the discussion of
[exception and flow commands](#exception-and-flow-commands) for more information
about this data type.

An exception is a [pseudo-map](#pseudo-map) with a `reason` field, which is in
turn a pseudo-map. The reason pseudo-map has has a `type` field identifying how
the exception was raised, and further fields depending on the type:

-   If the `type` field is `fail`, the exception was raised by the
    [fail](builtins.html#fail) command.

    In this case, the `content` field contains the argument to `fail`.

-   If the `type` field is `flow`, the exception was raised by one of the flow
    commands.

    In this case, the `name` field contains the name of the flow command.

-   If the `type` field is `pipeline`, the exception was a result of multiple
    commands in the same pipeline raising exceptions.

    In this case, the `exceptions` field contains the exceptions from the
    individual commands.

-   If the `type` field starts with `external-cmd/`, the exception was caused by
    one of several conditions of an external command. In this case, the
    following fields are available:

    -   The `cmd-name` field contains the name of the command.

    -   The `pid` field contains the PID of the command.

-   If the `type` field is `external-cmd/exited`, the external command exited
    with a non-zero status code. In this case, the `exit-status` field contains
    the exit status.

-   If the `type` field is `external-cmd/signaled`, the external command was
    killed by a signal. In this case, the following extra fields are available:

    -   The `signal-name` field contains the name of the signal.

    -   The `signal-number` field contains the numerical value of the signal, as
        a string.

    -   The `core-dumped` field is a boolean reflecting whether a core dump was
        generated.

-   If the `type` field is `external-cmd/stopped`, the external command was
    stopped. In this case, the following extra fields are available:

    -   The `signal-name` field contains the name of the signal.

    -   The `signal-number` field contains the numerical value of the signal, as
        a string.

    -   The `trap-cause` field contains the number indicating the trap cause.

Examples:

```elvish-transcript
~> put ?(fail foo)[reason]
▶ [&content=foo &type=fail]
~> put ?(return)[reason]
▶ [&name=return &type=flow]
~> put ?(false)[reason]
▶ [&cmd-name=false &exit-status=1 &pid=953421 &type=external-cmd/exited]
```

## File

There is no literal syntax for the file type. This type is returned by commands
such as [file:open](file.html#open) and [path:temp-file](path.html#temp-file).
It can be used as the target of a redirection rather than a filename.

A file object is a [pseudo-map](#pseudo-map) with fields `fd` (an int) and
`name` (a string). If the file is closed the fd will be -1.

## Function

A function encapsulates a piece of code that can be executed in an
[ordinary command](#ordinary-command), and takes its arguments and options.
Functions are first-class values; they can be kept in variables, used as
arguments, output on the value channel and embedded in other data structures.
Elvish comes with a set of **builtin functions**, and Elvish code can also
create **user-defined functions**.

**Note**: Unlike most programming languages, functions in Elvish do not have
return values. Instead, they can output values, which can be
[captured](#output-capture) later.

A **function literal**, or alternatively a **lambda**, evaluates to a
user-defined function. The literal syntax consists of an optional **signature
list**, followed by a [code chunk](#code-chunk) that defines the body of the
function.

Here is an example without a signature:

```elvish-transcript
~> f = { echo "Inside a lambda" }
~> put $f
▶ <closure 0x18a1a340>
```

One or more whitespace characters after `{` is required: Elvish relies on the
presence of whitespace to disambiguate function literals and
[braced lists](#braced-list).

**Note**: It is good style to put some whitespace before the closing `}` for
symmetry, but this is not required by the syntax.

Functions defined without a signature list do not accept any arguments or
options. To do so, write a signature list. Here is an example:

```elvish-transcript
~> f = [a b]{ put $b $a }
~> $f lorem ipsum
▶ ipsum
▶ lorem
```

There must be no space between `]` and `{`; otherwise Elvish will parse the
signature as a list, followed by a lambda without signature:

```elvish-transcript
~> put [a]{ nop }
▶ <closure 0xc420153d80>
~> put [a] { nop }
▶ [a]
▶ <closure 0xc42004a480>
```

Like in the left hand of assignments, if you prefix one of the arguments with
`@`, it becomes a **rest argument**, and its value is a list containing all the
remaining arguments:

```elvish-transcript
~> f = [a @rest]{ put $a $rest }
~> $f lorem
▶ lorem
▶ []
~> $f lorem ipsum dolar sit
▶ lorem
▶ [ipsum dolar sit]
~> f = [a @rest b]{ put $a $rest $b }
~> $f lorem ipsum dolar sit
▶ lorem
▶ [ipsum dolar]
▶ sit
```

You can also declare options in the signature. The syntax is `&name=default`
(like a map pair), where `default` is the default value for the option; the
value of the option will be kept in a variable called `name`:

```elvish-transcript
~> f = [&opt=default]{ echo "Value of $opt is "$opt }
~> $f
Value of $opt is default
~> $f &opt=foobar
Value of $opt is foobar
```

Options must have default values: Options should be **option**al.

If you call a function with too few arguments, too many arguments or unknown
options, an exception is thrown:

```elvish-transcript
~> [a]{ echo $a } foo bar
Exception: need 1 arguments, got 2
[tty], line 1: [a]{ echo $a } foo bar
~> [a b]{ echo $a $b } foo
Exception: need 2 arguments, got 1
[tty], line 1: [a b]{ echo $a $b } foo
~> [a b @rest]{ echo $a $b $rest } foo
Exception: need 2 or more arguments, got 1
[tty], line 1: [a b @rest]{ echo $a $b $rest } foo
~> [&k=v]{ echo $k } &k2=v2
Exception: unknown option k2
[tty], line 1: [&k=v]{ echo $k } &k2=v2
```

A user-defined function is a [pseudo-map](#pseudo-map). If `$f` is a
user-defined function, it has the following fields:

-   `$f[arg-names]` is a list containing the names of the arguments.

-   `$f[rest-arg]` is the index of the rest argument. If there is no rest
    argument, it is `-1`.

-   `$f[opt-names]` is a list containing the names of the options.

-   `$f[opt-defaults]` is a list containing the default values of the options,
    in the same order as `$f[opt-names]`.

-   `$f[def]` is a string containing the definition of the function, including
    the signature and the body.

-   `$f[body]` is a string containing the body of the function, without the
    enclosing brackets.

-   `$f[src]` is a map-like data structure containing information about the
    source code that the function is defined in. It contains the same value that
    the [src](builtin.html#src) function would output if called from the
    function.

# Variable

A variable is a named storage location for holding a value. The following
characters can be used in variable names (a subset of bareword characters)
without quoting:

A variable exist after it is declared (either explicitly using [`var`](#var) or
implicitly using the [legacy assignment form](#legacy-assignment-form)), and its
value may be mutated by further assignments. It can be [used](#variable-use) as
an expression or part of an expression.

**Note**: In most other shells, variables can map directly to environmental
variables: `$PATH` is the same as the `PATH` environment variable. This is not
the case in Elvish. Instead, environment variables are put in a dedicated `E:`
namespace; the environment variable `PATH` is known as `$E:PATH`. The `$PATH`
variable, on the other hand, does not exist initially, and if you have defined
it, only lives in a certain lexical scope within the Elvish interpreter.

You will notice that variables sometimes have a leading dollar `$`, and
sometimes not. The tradition is that they do when they are used for their
values, and do not otherwise (e.g. in assignment). This is consistent with most
other shells.

## Variable suffix

There are two characters that have special meanings and extra type constraints
when used as the suffix of a variable name:

-   If a variable name ends with `~`, it can only take callable values, which
    are functions and external commands. Such variables are consulted when
    resolving [ordinary commands](#ordinary-command). The default value is the
    builtin [`nop`](builtin.html#nop) command.

-   If a variable name ends with `:`, it can only take namespaces as values.
    They are used for accessing namespaced variables.

## Scoping rule

Elvish has lexical scoping. A file or an interactive prompt starts with a
top-level scope, and a [function literal](#function) introduce new lexical
scopes.

When you use a variable, Elvish looks for it in the current lexical scope, then
its parent lexical scope and so forth, until the outermost scope:

```elvish-transcript
~> x = 12
~> { echo $x } # $x is in the global scope
12
~> { y = bar; { echo $y } } # $y is in the outer scope
bar
```

If a variable is not in any of the lexical scopes, Elvish tries to resolve it in
the `builtin:` namespace, and if that also fails, fails with an error:

```elvish-transcript
~> echo $pid # builtin
36613
~> echo $nonexistent
Compilation error: variable $nonexistent not found
  [interactive], line 1:
    echo $nonexistent
```

Note that Elvish resolves all variables in a code chunk before starting to
execute any of it; that is why the error message above says _compilation error_.
This can be more clearly observed in the following example:

```elvish-transcript
~> echo pre-error; echo $nonexistent
Compilation error: variable $nonexistent not found
[tty], line 1: echo pre-error; echo $nonexistent
```

When you assign a variable, Elvish does a similar searching. If the variable
cannot be found, instead of causing an error, it will be created in the current
scope:

```elvish-transcript
~> x = 12
~> { x = 13 } # assigns to x in the global scope
~> echo $x
13
~> { z = foo } # creates z in the inner scope
~> echo $z
Compilation error: variable $z not found
[tty], line 1: echo $z
```

One implication of this behavior is that Elvish will not shadow your variable in
outer scopes.

There is a `local:` namespace that always refers to the current scope, and by
using it it is possible to force Elvish to shadow variables:

```elvish-transcript
~> x = 12
~> { local:x = 13; echo $x } # force shadowing
13
~> echo $x
12
```

After force shadowing, you can still access the variable in the outer scope
using the `up:` namespace, which always **skips** the innermost scope:

```elvish-transcript
~> x = 12
~> { local:x = 14; echo $x $up:x }
14 12
```

The `local:` and `up:` namespaces can also be used on unshadowed variables,
although they are not useful in those cases:

```elvish-transcript
~> foo = a
~> { echo $up:foo } # $up:foo is the same as $foo
a
~> { bar = b; echo $local:bar } # $local:bar is the same as $bar
b
```

It is not possible to refer to a specific outer scope.

You cannot create new variables in the `builtin:` namespace, although existing
variables in it can be assigned new values.

## Closure semantics

When a function literal refers to a variable in an outer scope, the function
will keep that variable alive, even if that variable is the local variable of an
outer function that that function has returned. This is called
[closure semantics](<https://en.wikipedia.org/wiki/Closure_(computer_programming)>),
because the function literal "closes" over the environment it is defined in.

In the following example, the `make-adder` function outputs two functions, both
referring to a local variable `$n`. Closure semantics means that:

1.  Both functions can continue to refer to the `$n` variable after `make-adder`
    has returned.

2.  Multiple calls to the `make-adder` function generates distinct instances of
    the `$n` variables.

```elvish-transcript
~> fn make-adder {
     n = 0
     put { put $n } { n = (+ $n 1) }
   }
~> getter adder = (make-adder)
~> $getter # $getter outputs $n
▶ 0
~> $adder # $adder increments $n
~> $getter # $getter and $setter refer to the same $n
▶ 1
~> getter2 adder2 = (make-adder)
~> $getter2 # $getter2 and $getter refer to different $n
▶ 0
~> $getter
▶ 1
```

Variables that get "captured" in closures are called **upvalues**; this is why
the pseudo-namespace for variables in outer scopes is called `up:`. When
capturing upvalues, Elvish only captures the variables that are used. In the
following example, `$m` is not an upvalue of `$g` because it is not used:

```elvish-transcript
~> fn f { m = 2; n = 3; put { put $n } }
~> g = (f)
```

This effect is not currently observable, but will become so when namespaces
[become introspectable](https://github.com/elves/elvish/issues/492).

# Expressions

Elvish has a few types of expressions. Some of those are new compared to most
other languages, but some are very similar.

Unlike most other languages, expressions in Elvish may evaluate to any number of
values. The concept of multiple values is distinct from a list of multiple
elements.

## Literal

Literals of [strings](#string), [lists](#list), [maps](#map) and
[functions](#function) all evaluate to one value of their corresponding types.
They are described in their respective sections.

## Variable use

A **variable use** expression is formed by a `$` followed by the name of the
variable. Examples:

```elvish-transcript
~> foo = bar
~> x y = 3 4
~> put $foo
▶ bar
~> put $x
▶ 3
```

If the variable name only contains the following characters (a subset of
bareword characters), the name can appear unquoted after `$` and the variable
use expression extends to the longest sequence of such characters:

-   ASCII letters (a-z and A-Z) and numbers (0-9);

-   The symbols `-_:~`. The colon `:` is special; it is normally used for
    separating namespaces or denoting namespace variables;

-   Non-ASCII codepoints that are printable, as defined by
    [unicode.IsPrint](https://godoc.org/unicode#IsPrint) in Go's standard
    library.

Alternatively, `$` may be followed immediately by a
[single-quoted string](https://elv.sh/ref/language.html#single-quoted-string) or
a [double-quoted string](https://elv.sh/ref/language.html#double-quoted-string),
in which cases the value of the string specifies the name of the variable.
Examples:

```elvish-transcript
~> "\n" = foo
~> put $"\n"
▶ foo
~> '!!!' = bar
~> put $'!!!'
▶ bar
```

Unlike other shells and other dynamic languages, local namespaces in Elvish are
statically checked. This means that referencing a nonexistent variable results
in a compilation error, which is triggered before any code is actually
evaluated:

```elvish-transcript
~> echo $x
Compilation error: variable $x not found
[tty], line 1: echo $x
~> f = { echo $x }
compilation error: variable $x not found
[tty 1], line 1: f = { echo $x }
```

If a variable contains a list value, you can add `@` before the variable name;
this evaluates to all the elements within the list. This is called **exploding**
the variable:

```elvish-transcript
~> li = [lorem ipsum foo bar]
~> put $li
▶ [lorem ipsum foo bar]
~> put $@li
▶ lorem
▶ ipsum
▶ foo
▶ bar
```

**Note**: Since variable uses have higher precedence than [indexing](#indexing),
this does not work for exploding a list that is an element of another list. For
doing that, and exploding the result of other expressions (such as an output
capture), use the builtin [all](builtin.html#all) command.)

## Output capture

An **output capture** expression is formed by putting parentheses `()` around a
[code chunk](#code-chunk). It redirects the output of the chunk into an internal
pipe, and evaluates to all the values that have been output.

```elvish-transcript
~> + 1 10 100
▶ 111
~> x = (+ 1 10 100)
~> put $x
▶ 111
~> put lorem ipsum
▶ lorem
▶ ipsum
~> x y = (put lorem ipsum)
~> put $x
▶ lorem
~> put $y
▶ ipsum
```

If the chunk outputs bytes, Elvish strips the last newline (if any), and split
them by newlines, and consider each line to be one string value:

```elvish-transcript
~> put (echo "a\nb")
▶ a
▶ b
```

Trailing carriage returns are also stripped from each line, which effectively
makes `\r\n` also valid line separators:

```elvish-transcript
~> put (echo "a\r\nb")
▶ a
▶ b
```

**Note 1**. Only the last newline is ever removed, so empty lines are preserved;
`(echo "a\n")` evaluates to two values, `"a"` and `""`.

**Note 2**. One consequence of this mechanism is that you can not distinguish
outputs that lack a trailing newline from outputs that have one; `(echo what)`
evaluates to the same value as `(print what)`. If such a distinction is needed,
use [`slurp`](builtin.html#slurp) to preserve the original bytes output.

If the chunk outputs both values and bytes, the values of output capture will
contain both value outputs and lines. However, the ordering between value output
and byte output might not agree with the order in which they happened:

```elvish-transcript
~> put (put a; echo b) # value order need not be the same as output order
▶ b
▶ a
```

## Exception capture

An **exception capture** expression is formed by putting `?()` around a code
chunk. It runs the chunk and evaluates to the exception it throws.

```elvish-transcript
~> fail bad
Exception: bad
Traceback:
  [interactive], line 1:
    fail bad
~> put ?(fail bad)
▶ ?(fail bad)
```

If there was no error, it evaluates to the special value `$ok`:

```elvish-transcript
~> nop
~> put ?(nop)
▶ $ok
```

Exceptions are booleanly false and `$ok` is booleanly true. This is useful in
`if` (introduced later):

```elvish-transcript
if ?(test -d ./a) {
  # ./a is a directory
}
```

**Note**: Exception captures do not affect the output of the code chunk. You can
combine output capture and exception capture:

```elvish
output = (error = ?(commands-that-may-fail))
```

## Braced list

A **braced list** consists of multiple expressions separated by whitespaces and
surrounded by braces (`{}`). There must be no space after the opening brace. A
braced list evaluates to whatever the expressions inside it evaluate to. Its
most typical use is grouping multiple values in a
[compound expression](#compounding). Example:

```elvish-transcript
~> put {a b}-{1 2}
▶ a-1
▶ a-2
▶ b-1
▶ b-2
```

It can also be used to affect the [order of evaluation](#order-of-evaluation).
Examples:

```elvish-transcript
~> put *
▶ foo
▶ bar
~> put *o
▶ foo
~> put {*}o
▶ fooo
▶ baro
```

**Note**: When used to affect the order of evaluation, braced lists are very
similar to parentheses in C-like languages.

**Note**: A braced list is an expression. It is a syntactical construct and not
a separate data structure.

Elvish currently also supports using commas to separate items in a braced list.
This will likely be removed in future, but it also means that literal commas
must be quoted right now.

## Indexing

An **indexing expression** is formed by appending one or more indices inside a
pair of brackets (`[]`) after another expression (the indexee). Examples:

```elvish-transcript
~> li = [foo bar]
~> put $li[0]
▶ foo
~> li = [[foo bar] quux]
~> put $li[0][0]
▶ foo
~> put [[foo bar]][0][0]
▶ foo
```

If the expression being indexed evaluates to multiple values, the indexing
operation is applied on each value. Example:

```elvish-transcript
~> put (put [foo bar] [lorem ipsum])[0]
▶ foo
▶ lorem
~> put {[foo bar] [lorem ipsum]}[0]
▶ foo
▶ lorem
```

If there are multiple index expressions, or the index expression evaluates to
multiple values, the indexee is indexed once for each of the index value.
Examples:

```elvish-transcript
~> put elv[0 2 0..2]
▶ e
▶ v
▶ el
~> put [lorem ipsum foo bar][0 2 0..2]
▶ lorem
▶ foo
▶ [lorem ipsum]
~> put [&a=lorem &b=ipsum &a..b=haha][a a..b]
▶ lorem
▶ haha
```

If both the indexee and index evaluate to multiple values, the results generated
from the first indexee appear first. Example:

```elvish-transcript
~> put {[foo bar] [lorem ipsum]}[0 1]
▶ foo
▶ bar
▶ lorem
▶ ipsum
```

## Compounding

A **compound expression** is formed by writing several expressions together with
no space in between. A compound expression evaluates to a string concatenation
of all the constituent expressions. Examples:

```elvish-transcript
~> put 'a'b"c" # compounding three string literals
▶ abc
~> v = value
~> put '$v is '$v # compounding one string literal with one string variable
▶ '$v is value'
```

When one or more of the constituent expressions evaluate to multiple values, the
result is all possible combinations:

```elvish-transcript
~> li = [foo bar]
~> put {a b}-$li[0 1]
▶ a-foo
▶ a-bar
▶ b-foo
▶ b-bar
```

The order of the combinations is determined by first taking the first value in
the leftmost expression that generates multiple values, and then taking the
second value, and so on.

## Tilde expansion

An unquoted tilde at the beginning of a compound expression triggers **tilde
expansion**. The remainder of this expression must be a string. The part from
the beginning of the string up to the first `/` (or the end of the word if the
string does not contain `/`), is taken as a user name; and they together
evaluate to the home directory of that user. If the user name is empty, the
current user is assumed.

In the following example, the home directory of the current user is
`/home/xiaq`, while that of the root user is `/root`:

```elvish-transcript
~> put ~
▶ /home/xiaq
~> put ~root
▶ /root
~> put ~/xxx
▶ /home/xiaq/xxx
~> put ~root/xxx
▶ /root/xxx
```

Note that tildes are not special when they appear elsewhere in a word:

```elvish-transcript
~> put a~root
▶ a~root
```

If you need them to be, use a [braced list](#braced-list):

```elvish-transcript
~> put a{~root}
▶ a/root
```

## Wildcard expansion

**Wildcard patterns** are expressions that contain **wildcards**. Wildcard
patterns evaluate to all filenames they match.

In examples in this section, we will assume that the current directory has the
following structure:

```
.x.conf
a.cc
ax.conf
foo.cc
d/
|__ .x.conf
|__ ax.conf
|__ y.cc
.d2/
|__ .x.conf
|__ ax.conf
```

Elvish supports the following wildcards:

-   `?` matches one arbitrary character except `/`. For example, `?.cc` matches
    `a.cc`;

-   `*` matches any number of arbitrary characters except `/`. For example,
    `*.cc` matches `a.cc` and `foo.cc`;

-   `**` matches any number of arbitrary characters including `/`. For example,
    `**.cc` matches `a.cc`, `foo.cc` and `b/y.cc`.

The following behaviors are default, although they can be altered by modifiers:

-   When the entire wildcard pattern has no match, an error is thrown.

-   None of the wildcards matches `.` at the beginning of filenames. For
    example:

    -   `?x.conf` does not match `.x.conf`;

    -   `d/*.conf` does not match `d/.x.conf`;

    -   `**.conf` does not match `d/.x.conf`.

Wildcards can be **modified** using the same syntax as indexing. For instance,
in `*[match-hidden]` the `*` wildcard is modified with the `match-hidden`
modifier. Multiple matchers can be chained like `*[set:abc][range:0-9]`. In
which case they are OR'ed together.

There are two kinds of modifiers:

**Global modifiers** apply to the whole pattern and can be placed after any
wildcard:

-   `nomatch-ok` tells Elvish not to throw an error when there is no match for
    the pattern. For instance, in the example directory `put bad*` will be an
    error, but `put bad*[nomatch-ok]` does exactly nothing.

-   `but:xxx` (where `xxx` is any filename) excludes the filename from the final
    result.

-   `type:xxx` (where `xxx` is a recognized file type from the list below). Only
    one type modifier is allowed. For example, to find the directories at any
    level below the current working directory: `**[type:dir]`.

    -   `dir` will match if the path is a directory.

    -   `regular` will match if the path is a regular file.

Although global modifiers affect the entire wildcard pattern, you can add it
after any wildcard, and the effect is the same. For example,
`put */*[nomatch-ok].cpp` and `put *[nomatch-ok]/*.cpp` do the same thing. On
the other hand, you must add it after a wildcard, instead of after the entire
pattern: `put */*.cpp[nomatch-ok]` unfortunately does not do the correct thing.
(This will probably be fixed.)

**Local modifiers** only apply to the wildcard it immediately follows:

-   `match-hidden` tells the wildcard to match `.` at the beginning of
    filenames, e.g. `*[match-hidden].conf` matches `.x.conf` and `ax.conf`.

    Being a local modifier, it only applies to the wildcard it immediately
    follows. For instance, `*[match-hidden]/*.conf` matches `d/ax.conf` and
    `.d2/ax.conf`, but not `d/.x.conf` or `.d2/.x.conf`.

-   Character matchers restrict the characters to match:

    -   Character sets, like `set:aeoiu`;

    -   Character ranges like `range:a-z` (including `z`) or `range:a~z`
        (excluding `z`);

    -   Character classes: `control`, `digit`, `graphic`, `letter`, `lower`,
        `mark`, `number`, `print`, `punct`, `space`, `symbol`, `title`, and
        `upper`. See the Is\* functions [here](https://godoc.org/unicode) for
        their definitions.

Note the following caveats:

-   Local matchers chained together in separate modifiers are OR'ed. For
    instance, `?[set:aeoiu][digit]` matches all files with the chars `aeoiu` or
    containing a digit.

-   Local matchers combined in the same modifier, such as `?[set:aeoiu digit]`,
    behave in a hard to explain manner. Do not use this form as **the behavior
    is likely to change in the future.**

-   Dots at the beginning of filenames always require an explicit
    `match-hidden`, even if the matcher includes `.`. For example,
    `?[set:.a]x.conf` does **not** match `.x.conf`; you have to
    `?[set:.a match-hidden]x.conf`.

-   Likewise, you always need to use `**` to match slashes, even if the matcher
    includes `/`. For example `*[set:abc/]` is the same as `*[set:abc]`.

## Order of evaluation

An expression can use a combination of indexing, tilde expansion, wildcard and
compounding. The order of evaluation is as follows:

1.  Literals, variable uses, output captures and exception captures and braced
    lists have the highest precedence and are evaluated first.

2.  Indexing has the next highest precedence and is then evaluated first.

3.  Expression compounding then happens. Tildes and wildcards are kept
    unevaluated.

4.  If the expression starts with a tilde, tilde expansion happens. If the tilde
    is followed by a wildcard, an exception is raised.

5.  If the expression contains any wildcard, wildcard expansion happens.

Here an example: in `~/$li[0 1]/*` (where `$li` is a list `[foo bar]`), the
expression is evaluated as follows:

1.  The variable use `$li` evaluates to the list `[foo bar]`.

2.  The indexing expression `$li[0]` evaluates to two strings `foo` and `bar`.

3.  Compounding the expression, the result is `~/foo/*` and `~/bar/*`.

4.  Tilde expansion happens; assuming that the user's home directory is
    `/home/elf`, the values are now `/home/elf/foo/*` and `/home/elf/bar/*`.

5.  Wildcard expansion happens, evaluating the expression to all the filenames
    within `/home/elf/foo` and `/home/elf/bar`. If any directory is empty or
    nonexistent, an exception is thrown.

To force a particular order of evaluation, group expressions using a
[braced list](#braced-list).

# Command forms

A **command form** is either an [ordinary command](#ordinary-command), a
[special command](#special-command) or an
[legacy assignment form](#legacy-assignment-form). All of three different types
have access to [IO ports](#io-ports), which can be modifed via
[redirections](#redirection).

When Elvish parses a command form, it applies the following process to decide
its type:

-   If the command form contains an unquoted equal sign surrounded by inline
    whitespaces, it is an ordinary assignment.

-   If the first expression in the command form contains a single string
    literal, and the string value matches one of the special commands, it is a
    special command.

-   Otherwise, it is an ordinary command.

## Ordinary command

An **ordinary command** form consists of a command head, and any number of
arguments and options.

The first expression in an ordinary command is the command **head**. If the head
is a single string literal it is subject to **static resolution**:

-   If a variable with name `head~` (where `head` is the value of the head)
    exists, the head will evaluate as if it is `$head~`; i.e., a function
    invocation.

-   Otherwise, the head will evaluate to an external command with the name
    `head`.

If the head is not a single string literal, it is evaluated as a normal
expression. The expression must evaluate to one value, and the value must be one
of the following:

-   A callable value: a function or external command.

-   A string containing at least one slash, in which case it is treated like an
    external command with the string value as its path.

Examples of commands using static resolution:

```elvish-transcript
~> put x # resolves to builtin function $put~
▶ x
~> f~ = { put 'this is f' }
~> f # resolves to user-defined function $f~
▶ 'this is f'
~> whoami # resolves to external command whoami
elf
```

Examples of commands using a dynamic callable head:

```elvish-transcript
~> $put~ x
▶ x
~> (external whoami)
elf
~> { put 'this is a lambda' }
▶ 'this is a lambda'
```

**Note**: The last command resembles a code block in C-like languages in syntax,
but is quite different under the hood: it works by defining a function on the
fly and calling it immediately.

Examples of commands using a dynamic string head:

```elvish-transcript
~> x = /bin/whoami
~> $x
elf
~> x = whoami
~> $x # dynamic strings can only used when containing slash
Exception: bad value: command must be callable or string containing slash, but is string
[tty 10], line 1: $x
```

The definition of barewords is relaxed when parsing the head, and includes `<`,
`>`, and `*`. These are all names of numeric builtins:

```elvish-transcript
~> < 3 5 # less-than
▶ $true
~> > 3 5 # greater-than
▶ $false
~> * 3 5 # multiplication
▶ 15
```

**Arguments** and **options** can be supplied to commands. Arguments are
arbitrary words, while options have exactly the same syntax as key-value pairs
in [map literals](#map). They are separated by inline whitespaces and may be
intermixed:

```elvish-transcript
~> echo &sep=, a b c # &seq=, is an option; a b c are arguments
a,b,c
~> echo a b &sep=, c # same, with the option mixed within arguments
a,b,c
```

**Note**: Since options have the same syntax as key-value pairs in maps, `&key`
is equivalent to `&key=$true`:

```elvish-transcript
~> fn f [&opt=$false]{ put $opt }
~> f &opt
▶ $true
```

## Special command

A **special command** form has the same syntax with an ordinary command, but how
it is executed depends on the command head. See
[special commands](#special-commands).

## Legacy assignment form

If any argument in a command form is an unquoted equal sign (`=`), the command
form is treated as an assignment form: the arguments to the left of `=`,
including the head, are treated as lvalues, and the arguments to the right of
`=` are treated as values to assign to those lvalues.

If any lvalue refers to a variable that doesn't yet exist, it is created first.

This is a legacy syntax that will be deprecated in future. Use the [`var`](#var)
special command to declare variables, and the [`set`](#set) special command set
the values of variables.

## Temporary assignment

You can prepend any command form with **temporary assignments**, which gives
variables temporarily values during the execution of that command.

In the following example, `$x` and `$y` are temporarily assigned 100 and 200:

```elvish-transcript
~> x y = 1 2
~> x=100 y=200 + $x $y
▶ 300
~> echo $x $y
1 2
```

In contrary to normal assignments, there should be no whitespaces around the
equal sign `=`. To have multiple variables in the left-hand side, use braces:

```elvish-transcript
~> x y = 1 2
~> fn f { put 100 200 }
~> {x,y}=(f) + $x $y
▶ 300
```

If you use a previously undefined variable in a temporary assignment, its value
will become the empty string after the command finishes. This behavior will
likely change; don't rely on it.

Since ordinary assignments are also command forms, they can also be prepended
with temporary assignments:

```elvish-transcript
~> x=1
~> x=100 y = (+ 133 $x)
~> put $x $y
▶ 1
▶ 233
```

Temporary assignments must all appear at the beginning of the command form. As
soon as something that is not a temporary assignments is parsed, Elvish no
longer parses temporary assignments. For instance, in `x=1 echo x=1`, the second
`x=1` is not a temporary assignment, but a bareword.

**Note**: Elvish's behavior differs from bash (or zsh) in one important place.
In bash, temporary assignments to variables do not affect their direct
appearance in the command:

```sh-transcript
bash-4.4$ x=1
bash-4.4$ x=100 echo $x
1
```

**Note**: Elvish currently supports using the syntax of temporary assignments
for ordinary assignments, when they are not followed by a command form; for
example, `a=x` behaves like an ordinary assignment `a = x`. This will likely go
away; don't rely on it.

## IO ports

A command have access to a number of **IO ports**. Each IO port is identified by
a number starting from 0, and combines a traditional file object, which conveys
bytes, and a **value channel**, which conveys values.

Elvish starts with 3 IO ports at the top level with special significance for
commands:

-   Port 0, known as standard input or stdin, and is used as the default input
    port by builtin commands.

-   Port 1, known as standard output or stdout, and is used as the default
    output port by builtin commands.

-   Port 2, known as standard error or stderr, is currently not special for
    builtin commands, but usually has special significance for external
    commands.

Value channels are typically created by a [pipeline](#pipeline), and used to
pass values between commands in the same pipeline. At the top level, they are
initialized with special values:

-   The value channel for port 0 never produces any values when read.

-   The value channels for port 1 and 2 are special channels that forward the
    values written to them to their file counterparts. Each value is put on a
    separate line, with a prefix controlled by
    [`$value-out-indicator`](builtin.html#value-out-indicator). The default
    prefix is `▶` followed by a space.

When running an external command, the file object from each port is used to
create its file descriptor table. Value channels only work inside the Elvish
process, and are not accessible to external commands.

IO ports can be modified with [redirections](#redirection) or by
[pipelines](#pipeline).

## Redirection

<!-- TODO: Describe the syntax and behavior more formally. -->

A **redirection** modifies the IO ports a command operate with. There are
several variants.

A **file redirection** opens a file and associates it with an IO port. The
syntax consists of an optional destination IO port (like `2`), a redirection
operator (like `>`) and a filename (like `error.log`):

-   The **destination IO port** determines which IO port to modify. It can be
    given either as the number of the IO port, or one of `stdin`, `stdout` and
    `stderr`, which are equivalent to 0, 1 and 2 respectively.

    The destination IO port can be omitted, in which case it is inferred from
    the redirection operator.

    When the destination IO port is given, it must precede the redirection
    operator directly, without whitespaces in between; if there are whitespaces,
    otherwise Elvish will parse it as an argument instead.

-   The **redirection operator** determines the mode to open the file, and the
    destination IO port if it is not explicitly specified.

-   The **filename** names the file to open.

Possible redirection operators and their default FDs are:

-   `<` for reading. The default IO port is 0 (stdin).

-   `>` for writing. The default IO port is 1 (stdout).

-   `>>` for appending. The default IO port is 1 (stdout).

-   `<>` for reading and writing. The default IO port is 1 (stdout).

Examples:

```elvish-transcript
~> echo haha > log
~> cat log
haha
~> cat < log
haha
~> ls --bad-arg 2> error
Exception: ls exited with 2
Traceback:
  [interactive], line 1:
    ls --bad-arg 2> error
~> cat error
/bin/ls: unrecognized option '--bad-arg'
Try '/bin/ls --help' for more information.
```

IO ports modified by file redirections do not support value channels. To be more
exact:

-   A file redirection using `<` sets the value channel to one that never
    produces any values.

-   A file redirection using `>`, `>>` or `<>` sets the value channel to one
    that throws an exception when written to.

Examples:

```elvish-transcript
~> put foo > file # will truncate file if it exists
Exception: port has no value output
[tty 2], line 1: put foo > file
~> echo content > file
~> only-values < file
~> # previous command produced nothing
```

Redirections can also be used for closing or duplicating IO ports. Instead of
writing a filename, use `&fd` (where `fd` is a number, or any of `stdin`,
`stdout` and `stderr`) for duplicating, or `&-` for closing. In this case, the
redirection operator only determines the default destination FD (and is totally
irrevelant if a destination IO port is specified). Examples:

```elvish-transcript
~> date >&-
date: stdout: Bad file descriptor
Exception: date exited with 1
[tty 3], line 1: date >&-
~> put foo >&-
Exception: port has no value output
[tty 37], line 1: put foo >&-
```

If you have multiple related redirections, they are applied in the order they
appear. For instance:

```elvish-transcript
~> fn f { echo out; echo err >&2 } # echoes "out" on stdout, "err" on stderr
~> f >log 2>&1 # use file "log" for stdout, then use (changed) stdout for stderr
~> cat log
out
err
```

Redirections may appear anywhere in the command, except at the beginning, (this
may be restricted in future). It's usually good style to write redirections at
the end of command forms.

# Special commands

**Special commands** obey the same syntax rules as normal commands, but have
evaluation rules that are custom to each command. Consider the following
example:

```elvish-transcript
~> or ?(echo x) ?(echo y) ?(echo z)
x
▶ $ok
```

In the example, the `or` command first evaluates its first argument, which has
the value `$ok` (a truish value) and the side effect of outputting `x`. Due to
the custom evaluation rule of `or`, the rest of the arguments are not evaluated.

If `or` were a normal command, the code above is still syntactically correct.
However, Elvish would then evaluate all its arguments, with the side effect of
outputting `x`, `y` and `z`, before calling `or`.

## Declaring variables: `var` {#var}

The `var` special command declares local variables. It takes any number of
unqualified variable names (without the leading `$`). The variables will start
out having value `$nil`. Examples:

```elvish-transcript
~> var a
~> put $a
▶ $nil
~> var foo bar
~> put $foo $bar
▶ $nil
▶ $nil
```

To set alternative initial values, add an unquoted `=` and the initial values.
Examples:

```elvish-transcript
~> var a b = foo bar
~> put $a $b
▶ foo
▶ bar
```

Similar to [`set`](#set), at most one of variables may be prefixed with `@` to
function as a rest variable.

When declaring a variable that already exists, the existing variable is
shadowed. The shadowed variable may still be accessed indirectly if it is
referenced by a function. Example:

```elvish-transcript
~> var x = old
~> fn f { put $x }
~> var x = new
~> put $x
▶ new
~> f
▶ old
```

## Setting the value of variables or elements: `set` {#set}

The `set` special command sets the value of variables or elements.

It takes any number of **lvalues** (which refer to either variables or
elements), followed by an equal sign (`=`) and any number of expressions. The
equal sign must appear unquoted, as a single argument.

An **lvalue** is one of the following:

-   A variable name (without `$`).

-   A variable name prefixed with `@`, for packing a variable number of values
    into a list and assigning to the variable.

    This variant is called a **rest variable**. There could be at most one rest
    variable.

    **Note**: Schematically this is the reverse operation of exploding a
    variable when [using](#variable-use) it, which is why they share the `@`
    sign.

-   A variable name followed by one or more indices in brackets (`[]`), for
    assigning to an element.

The number of values the expressions evaluate to and lvalues must be compatible.
To be more exact:

-   If there is no rest variable, the number of values and lvalues must match
    exactly.

-   If there is a rest variable, the number of values should be at least the
    number of lvalues minus one.

All the variables to set must already exist; use the [`var`](#var) special
command to declare new variables.

Examples:

```elvish-transcript
~> var x y z
~> set x = foo
~> put $x
▶ foo
~> x y = lorem ipsum
~> put $x $y
▶ lorem
▶ ipsum
~> set x @y z = a b
~> put $x $y $z
▶ a
▶ []
▶ b
~> set x @y z = a b c d
~> put $x $y $z
▶ a
▶ [b c]
▶ d
~> set y[0] = foo
~> put $y
▶ [foo c]
```

If the variable name contains any character that may not appear unquoted in
[variable use expressions](#variable-use), it must be quoted even if it is
otherwise a valid bareword:

```elvish-transcript
~> var 'a/b'
~> set a/b = foo
compilation error: lvalue must be valid literal variable names
[tty 23], line 1: a/b = foo
~> set 'a/b' = foo
~> put $'a/b'
▶ foo
```

Lists and maps in Elvish are immutable. As a result, when assigning to the
element of a variable that contains a list or map, Elvish does not mutate the
underlying list or map. Instead, Elvish creates a new list or map with the
mutation applied, and assigns it to the variable. Example:

```elvish-transcript
~> var li = [foo bar]
~> var li2 = $li
~> set li[0] = lorem
~> put $li $li2
▶ [lorem bar]
▶ [foo bar]
```

## Deleting variables or elements: `del` {#del}

The `del` special command can be used to delete variables or map elements.
Operands should be specified without a leading dollar sign, like the left-hand
side of assignments.

Example of deleting variable:

```elvish-transcript
~> x = 2
~> echo $x
2
~> del x
~> echo $x
Compilation error: variable $x not found
[tty], line 1: echo $x
```

If the variable name contains any character that cannot appear unquoted after
`$`, it must be quoted, even if it is otherwise a valid bareword:

```elvish-transcript
~> 'a/b' = foo
~> del 'a/b'
```

Deleting a variable does not affect closures that have already captured it; it
only removes the name. Example:

```elvish-transcript
~> x = value
~> fn f { put $x }
~> del x
~> f
▶ value
```

Example of deleting map element:

```elvish-transcript
~> m = [&k=v &k2=v2]
~> del m[k2]
~> put $m
▶ [&k=v]
~> l = [[&k=v &k2=v2]]
~> del l[0][k2]
~> put $l
▶ [[&k=v]]
```

## Logics: `and` and `or` {#and-or}

The `and` special command evaluates its arguments from left to right; as soon as
a booleanly false value is obtained, it outputs the value and stops. When given
no arguments, it outputs `$true`.

The `or` special command is the same except that it stops when a booleanly true
value is obtained. When given no arguments, it outpus `$false`.

## Condition: `if` {#if}

**TODO**: Document the syntax notation, and add more examples.

Syntax:

```elvish-transcript
if <condition> {
    <body>
} elif <condition> {
    <body>
} else {
    <else-body>
}
```

The `if` special command goes through the conditions one by one: as soon as one
evaluates to a booleanly true value, its corresponding body is executed. If none
of conditions are booleanly true and an else body is supplied, it is executed.

The condition part is an expression, not a command like in other shells.
Example:

```elvish
fn tell-language [fname]{
    if (has-suffix $fname .go) {
        echo $fname" is a Go file!"
    } elif (has-suffix $fname .c) {
        echo $fname" is a C file!"
    } else {
        echo $fname" is a mysterious file!"
    }
}
```

The condition part must be syntactically a single expression, but it can
evaluate to multiple values, in which case they are and'ed:

```elvish
if (put $true $false) {
    echo "will not be executed"
}
```

If the expression evaluates to 0 values, it is considered true, consistent with
how `and` works.

Tip: a combination of `if` and `?()` gives you a semantics close to other
shells:

```elvish
if ?(test -d .git) {
    # do something
}
```

However, for Elvish's builtin predicates that output values instead of throw
exceptions, the output capture construct `()` should be used.

## Conditional loop: `while` {#while}

Syntax:

```elvish-transcript
while <condition> {
    <body>
} else {
    <else-body>
}
```

Execute the body as long as the condition evaluates to a booleanly true value.

The else body, if present, is executed if the body has never been executed (i.e.
the condition evaluates to a booleanly false value in the very beginning).

## Iterative loop: `for` {#for}

Syntax:

```elvish-transcript
for <var> <container> {
    <body>
} else {
    <body>
}
```

Iterate the container (e.g. a list). In each iteration, assign the variable to
an element of the container and execute the body.

The else body, if present, is executed if the body has never been executed (i.e.
the iteration value has no elements).

## Exception control: `try` {#try}

(If you just want to capture the exception, you can use the more concise
exception capture construct `?()` instead.)

Syntax:

```elvish-transcript
try {
    <try-block>
} except except-varname {
    <except-block>
} else {
    <else-block>
} finally {
    <finally-block>
}
```

Only `try` and `try-block` are required. This control structure behaves as
follows:

1.  The `try-block` is always executed first.

2.  If `except` is present and an exception occurs in `try-block`, it is caught
    and stored in `except-varname`, and `except-block` is then executed.
    Example:

    ```elvish-transcript
    ~> try { fail bad } except e { put $e }
    ▶ ?(fail bad)
    ```

    Note that if `except` is not present, exceptions thrown from `try` are not
    caught: for instance, `try { fail bad }` throws `bad`; it is equivalent to a
    plain `fail bad`.

    Note that the word after `except` names a variable, not a matching
    condition. Exception matching is not supported yet. For instance, you may
    want to only match exceptions that were created with `fail bad` with
    `except bad`, but in fact this creates a variable `$bad` that contains
    whatever exception was thrown.

3.  If no exception occurs and `else` is present, `else-block` is executed.
    Example:

    ```elvish-transcript
    ~> try { nop } else { echo well }
    well
    ```

4.  If `finally-block` is present, it is executed. Examples:

    ```elvish-transcript
    ~> try { fail bad } finally { echo final }
    final
    Exception: bad
    Traceback:
      [tty], line 1:
        try { fail bad } finally { echo final }
    ~> try { echo good } finally { echo final }
    good
    final
    ```

5.  If the exception was not caught (i.e. `except` is not present), it is
    rethrown.

Exceptions thrown in blocks other than `try-block` are not caught. If an
exception was thrown and either `except-block` or `finally-block` throws another
exception, the original exception is lost. Examples:

```elvish-transcript
~> try { fail bad } except e { fail worse }
Exception: worse
Traceback:
  [tty], line 1:
    try { fail bad } except e { fail worse }
~> try { fail bad } except e { fail worse } finally { fail worst }
Exception: worst
Traceback:
  [tty], line 1:
    try { fail bad } except e { fail worse } finally { fail worst }
```

## Function definition: `fn` {#fn}

Syntax:

```elvish-transcript
fn <name> <lambda>
```

Define a function with a given name. The function behaves in the same way to the
lambda used to define it, except that it "captures" `return`. In other words,
`return` will fall through lambdas not defined with `fn`, and continues until it
exits a function defined with `fn`:

```elvish-transcript
~> fn f {
     { echo a; return }
     echo b # will not execute
   }
~> f
a
~> {
     f
     echo c # executed, because f "captures" the return
   }
a
c
```

**TODO**: Find a better way to describe this. Hopefully the example is
illustrative enough, though.

The lambda may refer to the function being defined. This makes it easy to define
recursive functions:

```elvish-transcript
~> fn f [n]{ if (== $n 0) { put 1 } else { * $n (f (- $n 1)) } }
~> f 3
▶ (float64 6)
```

Under the hood, `fn` defines a variable with the given name plus `~` (see
[variable suffix](#variable-suffix)). Example:

```elvish-transcript
~> fn f { echo hello from f }
~> var v = $f~
~> $v
hello from f
```

# Pipeline

A **pipeline** is formed by joining one or more commands together with the pipe
sign (`|`).

For each pair of adjacent commands `a | b`, the standard output of the left-hand
command `a` (IO port 1) is connected to the standard input (IO port 0) of the
right-hand command `b`. Both the file and the value channel are connected, even
if one of them is not used.

Elvish may have internal buffering for both the file and the value channel, so
`a` may be able to write bytes or values even if `b` is not reading them. The
exact buffer size is not specified.

Command redirections are applied before the connection happens. For instance,
the following writes `foo` to `a.txt` instead of the output:

```elvish-transcript
~> echo foo > a.txt | cat
~> cat a.txt
foo
```

A pipeline runs all of its command in parallel, and terminates when all of the
commands have terminated.

## Pipeline exception

If one or more command in a pipeline throws an exception, the other commands
will continue to execute as normal. After all commands finish execution, an
exception is thrown, the value of which depends on the number of commands that
have thrown an exception:

-   If only one command has thrown an exception, that exception is rethrown.

-   If more than one commands have thrown exceptions, a "composite exception",
    containing information all exceptions involved, is thrown.

If a command threw an exception because it tried to write output when the next
command has terminated, that exception is suppressed when it is propagated to
the pipeline.

For example, the `put` command throws an exception when trying to write to a
closed pipe, so the following loop will terminate with an exception:

```elvish-transcript
~> while $true { put foo } > &-
Exception: port has no value output
[tty 9], line 1: while $true { put foo } > &-
```

However, if it appears in a pipeline before `nop`, the entire pipeline will not
throw an exception:

```elvish-transcript
~> while $true { put foo } | nop
~> # no exception thrown from previous line
```

Internally, the `put foo` command still threw an exception, but since that
exception was trying to write to output when `nop` already terminated, that
exception was suppressed by the pipeline.

This can be more clearly observed with the following code:

```elvish-transcript
~> var r = $false
~> { while $true { put foo }; set r = $true } | nop
~> put $r
▶ $false
```

The same mechanism works for builtin commands that write to the byte output:

```elvish-transcript
~> var r = $false
~> { while $true { echo foo }; set r = $true } | nop
~> put $r
▶ $false
```

On UNIX, if an external command was terminated by SIGPIPE, and Elvish detected
that it terminated after the next command in the pipeline, such exceptions will
also be suppressed by the pipeline. For example, the following pipeline does not
throw an exception, despite the `yes` command being killed by SIGPIPE:

```elvish-transcript
~> yes | head -n1
y
```

## Background pipeline

Adding an ampersand `&` to the end of a pipeline will cause it to be executed in
the background. In this case, the rest of the code chunk will continue to
execute without waiting for the pipeline to finish. Exceptions thrown from the
background pipeline do not affect the code chunk that contains it.

When a background pipeline finishes, a message is printed to the terminal if the
shell is interactive.

# Code Chunk

A **code chunk** is formed by joining zero or more pipelines together,
separating them with either newlines or semicolons.

Pipelines in a code chunk are executed in sequence. If any pipeline throws an
exception, the execution of the whole code chunk stops, propagating that
exception.

# Exception and Flow Commands

Exceptions have similar semantics to those in Python or Java. They can be thrown
with the [fail](builtin.html#fail) command and caught with either exception
capture `?()` or the `try` special command.

If an external command exits with a non-zero status, Elvish treats that as an
exception.

Flow commands -- `break`, `continue` and `return` -- are ordinary builtin
commands that raise special "flow control" exceptions. The `for`, `while`, and
`peach` commands capture `break` and `continue`, while `fn` modifies its closure
to capture `return`.

One interesting implication is that since flow commands are just ordinary
commands you can build functions on top of them. For instance, this function
`break`s randomly:

```elvish
fn random-break {
  if eq (randint 2) 0 {
    break
  }
}
```

The function `random-break` can then be used in for-loops and while-loops.

Note that the `return` flow control exception is only captured by functions
defined with `fn`. It falls through ordinary lambdas:

```elvish
fn f {
  {
    # returns f, falling through the innermost lambda
    return
  }
}
```

# Namespaces and Modules

Like other modern programming languages, but unlike traditional shells, Elvish
has a **namespace** mechanism for preventing name collisions.

## Syntax

Prepend `namespace:` to command names and variable names to specify the
namespace. The following code

```elvish
e:echo $E:PATH
```

uses the `echo` command from the `e:` namespace and the `PATH` variable from the
`E:` namespace. The colon is considered part of the namespace name.

Namespaces may be nested; for example, calling `edit:location:start` first finds
the `edit:` namespace, and then the `location:` namespace inside it, and then
call the `start` function within the nested namespace.

## Special namespaces

The following namespaces have special meanings to the language:

-   `local:` and `up:` refer to lexical scopes, and have been documented above.

-   `e:` refers to externals. For instance, `e:ls` refers to the external
    command `ls`.

    Most of the time you can rely on static resolution rules of
    [ordinary commands](#ordinary-command) and do not need to use this
    explicitly, unless a function defined by you (or an Elvish builtin) shadows
    an external command.

-   `E:` refers to environment variables. For instance, `$E:USER` is the
    environment variable `USER`.

    This **is** always needed, because unlike command resolution, variable
    resolution does not fall back onto environment variables.

-   `builtin:` refers to builtin functions and variables.

    You don't need to use this explicitly unless you have defined names that
    shadows builtin counterparts.

## Modules

Apart from the special namespaces, the most common usage of namespaces is to
reference modules, reusable pieces of code that are either shipped with Elvish
itself or defined by the user.

### Importing modules with `use`

Modules are imported using the `use` special command, which accepts a **module
spec** and an optional alias:

```elvish
use $spec $alias?
```

Both the module spec and the alias must appear as a single string literal.

The module spec specifies which module to import, and the alias, if given,
specifies the namespace to import the module under. By default, the namespace is
derived from the module spec, by taking the part after the last slash.

Examples:

```elvish
use str # imports the "str" module as "str:"
use a/b/c # imports the "a/b/c" module as "c:"
use a/b/c foo # imports the "a/b/c" module as "foo:"
```

### Pre-defined modules

Elvish's standard library provides the following pre-defined modules that can be
imported by the `use` command:

-   [edit](edit.html) is only available in interactive mode. As a special case
    it does not need importing via `use`, but this may change in the future.
-   [epm](epm.html)
-   [math](math.html)
-   [path](path.html)
-   [platform](platform.html)
-   [re](re.html)
-   [readline-binding](readline-binding.html)
-   [store](store.html)
-   [str](str.html)
-   [unix](unix.html) is only available on UNIX-like platforms (see
    [`$platform:is-unix`](platform.html#platformis-unix))

### User-defined modules

You can define your own modules in Elvish by putting them under `~/.elvish/lib`
and giving them a `.elv` extension. For instance, to define a module named `a`,
store it in `~/.elvish/lib/a.elv`:

```elvish-transcript
~> cat ~/.elvish/lib/a.elv
echo "mod a loading"
fn f {
  echo "f from mod a"
}
```

This module can now be imported by `use a`:

```elvish-transcript
~> use a
mod a loading
~> a:f
f from mod a
```

Similarly, a module defined in `~/.elvish/lib/x/y/z.elv` can be imported by
`use x/y/z`:

```elvish-transcript
~> cat .elvish/lib/x/y/z.elv
fn f {
  echo "f from x/y/z"
}
~> use x/y/z
~> z:f
f from x/y/z
```

In general, a module defined in namespace will be the same as the file name
(without the `.elv` extension).

There is experimental support for importing modules written in Go. See the
[project repository](https://github.com/elves/elvish) for details.

### Circular dependencies

Circular dependencies are allowed but has an important restriction. If a module
`a` contains `use b` and module `b` contains `use a`, the top-level statements
in module `b` will only be able to access variables that are defined before the
`use b` in module `a`; other variables will be `$nil`.

On the other hand, functions in module `b` will have access to bindings in
module `a` after it is fully evaluated.

Examples:

```elvish-transcript
~> cat a.elv
var before = before
use ./b
var after = after
~> cat b.elv
use ./a
put $a:before $a:after
fn f { put $a:before $a:after }
~> use ./a
▶ before
▶ $nil
~> use ./b
~> b:f
▶ before
▶ after
```

Note that this behavior can be different depending on whether the REPL imports
`a` or `b` first. In the previous example, if the REPL imports `b` first, it
will have access to all the variables in `a`:

```elvish-transcript
~> use ./b
▶ before
▶ after
```

**Note**: Elvish caches imported modules. If you are trying this locally, run a
fresh Elvish instance with `exec` first.

When you do need to have circular dependencies, it is best to avoid using
variables from the modules in top-level statements, and only use them in
functions.

### Relative imports

The module spec may being with `./` or `../`, which introduce **relative
imports**. When `use` is invoked from a file, this will import the file relative
to the location of the file. When `use` is invoked at the interactive prompt,
this will import the file relative to the current working directory.

### Scoping of imports

Namespace imports are lexically scoped. For instance, if you `use` a module
within an inner scope, it is not available outside that scope:

```elvish
{
    use some-mod
    some-mod:some-func
}
some-mod:some-func # not valid
```

The imported modules themselves are also evaluated in a separate scope. That
means that functions and variables defined in the module does not pollute the
default namespace, and vice versa. For instance, if you define `ls` as a wrapper
function in `rc.elv`:

```elvish
fn ls [@a]{
    e:ls --color=auto $@a
}
```

That definition is not visible in module files: `ls` will still refer to the
external command `ls`, unless you shadow it in the very same module.

### Re-importing

Modules are cached after one import. Subsequent imports do not re-execute the
module; they only serve the bring it into the current scope. Moreover, the cache
is keyed by the path of the module, not the name under which it is imported. For
instance, if you have the following in `~/.elvish/lib/a/b.elv`:

```elvish
echo importing
```

The following code only prints one `importing`:

```elvish
{ use a/b }
use a/b # only brings mod into the lexical scope
```

As does the following:

```elvish
use a/b
use a/b alias
```
