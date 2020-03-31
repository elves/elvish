<!-- toc number-sections -->

# Introduction

This document describes the Elvish programming language. It tries to be both a
specification and an advanced tutorial; if it turns out to be impossible to do
these two things at the same time, this document will evolve to a formal
specification, and more readable tutorials will be created.

Examples for one construct might use constructs that have not yet been
introduced, so some familiarity with the language is assumed. If you are new to
Elvish, start with the [learning materials](../learn/).

**Note to the reader**. Like Elvish itself, this document is a work in progress.
Some materials are missing, and some are documented sparingly. If you have found
something that should be improved -- even if there is already a "TODO" for it --
please feel free to ask on any of the chat channels advertised on the
[homepage](..). Some developer will explain to you, and then update the
document. Question-driven documentation :)

# Syntax Convention

Elvish source code must be UTF-8-encoded. In this document, **character** is a
synonym of [Unicode codepoint](https://en.wikipedia.org/wiki/Code_point) or its
UTF-8 encoding.

Also like most shells, Elvish uses whitespaces -- instead of commas, periods or
semicolons -- to separate constructs. In this document, an **inline whitespace**
is any of:

-   A space (ASCII 0x20) or tab (ASCII 0x9, `"\t"`);

-   A comment: starting with `#` and ending before the next newline or end of
    file;

-   Line continuation: a backslash followed by a newline.

A **whitespace** is either an **inline whitespace** or a newline (`"\n"`).

Like most shells, Elvish has a syntax structure that can be divided into two
levels: a **statement** level and an **expression** level. For instance, on the
expression level, `"echo"` is a quoted string that evaluates to `echo`; but on
the statement level, it is a command that outputs an empty line. This
distinction is often clear from the context and sometime the verb used. A
statements **executes** to produce side effects; an expression **evaluates** to
some values. (The traditional terms for the two levels are "commands" and
"words", but those terms are quite ambiguous.)

# Data types

## String

The most common data structure in shells is the string. String literals can be
quoted or unquoted (barewords).

### Quoted

There are two types of quoted strings in Elvish, single-quoted strings and
double-quoted strings.

In single-quoted strings, all characters represent themselves, except single
quotes, which need to be doubled. For instance, `'*\'` evaluates to `*\`, and
`'it''s'` evaluates to `it's`.

In double-quoted strings, the backslash `\` introduces a **escape sequence**.
For instance, `"\n"` evaluates to a newline; `"\\"` evaluates to a backslash;
invalid escape sequences like `"\*"` result in a syntax error.

**TODO**: Document the full list of supported escape sequences.

Unlike most other shells, double-quoted strings do not support interpolation.
For instance, `"$USER"` simply evaluates to the string `$USER`. To get a similar
effect, simply concatenate strings: instead of `"my name is $name"`, write
`"my name is "$name`. Under the hood this is a
[compound expression](#compound-expression-and-braced-lists).

### Barewords

If a string only consists of bareword characters, it can be written without any
quote; this is called a **bareword**. Examples are `a.txt`, `long-bareword`, and
`/usr/local/bin`. The set of bareword characters include:

-   ASCII letters (a-z and A-Z) and numbers (0-9);

-   The symbols `-_:%+,./@!`;

-   Non-ASCII codepoints that are printable, as defined by
    [unicode.IsPrint](https://godoc.org/unicode#IsPrint) in Go's standard
    library.

The following are bareword characters depending on their position:

-   The tilde `~`, unless it appears at the beginning of a compound expression,
    in which case it is subject to [tilde expansion](#tilde-expansion);

-   The equal sign `=`, unless it is used for terminating [map keys](#map) or
    [option keys](#arguments-and-options), or denoting
    [assignments](#assignment) or
    [temporary assignments](#temporary-assignment).

Unlike traditional shells, an unquoted backslash `\` does not escape
metacharacters; use quoted strings instead. For instance, to echo a star, write
`echo "*"` or `echo '*'`, not `echo \*`. Unquote backslashes are now only used
in line continuations; their use elsewhere is reserved will cause a syntax
error.

### Notes

The three syntaxes above all evaluate to strings, and they are interchangeable.
For instance, `xyz`, `'xyz'` and `"xyz"` are different syntaxes for the same
string, and they are always equivalent with the exception of
**escape sequences** as documented above.

## Number

Elvish has a double-precision floating point number type that can be constructed
with the `float64` builtin. The builtin takes a single argument, which should be
either another `float64` value, or a string in the following formats (examples
below all express the same value):

* Decimal notation, e.g. `10`.
  
* Hexadecimal notation, e.g. `0xA`.
  
* Octal notation, e.g. `0o12`.
  
* Binary notation, e.g. `0b1010`.
  
* Floating point notation, e.g. `10.0`.
  
* Scientific notation, e.g. `1.0e1`.

The following special floating point values are also supported: `+Inf`, `-Inf`
and `NaN`.

The `float64` builtin is case-insensitive.

A `float64` data type can be converted to a string using `(to-string
$number)`. The resulting string is guaranteed to result in the same
value when converted back to a `float64`. Most of the time you won't
need to perform this explicit conversion. Elvish will implicitly make
the conversion when running external commands and many of the builtins
(where the distinction is not important).

You usually do not need to use `float64` values explicitly;
see the discussion of [Commands That Operate On
Numbers](./builtin.html#commands-that-operate-on-numbers).

## Exception

Elvish has an exception data type, but it does not have a literal
syntax for that type. See the discussion of [exception and flow
commands](./language.html#exception-and-flow-commands) for more
information about this data type.

## List

Lists are surround by square brackets `[ ]`, with elements separated by
whitespace. They are one of the basic container types in Elvish. Examples:

```elvish-transcript
~> put [lorem ipsum]
▶ [lorem ipsum]
~> put [lorem
        ipsum
        foo
        bar]
▶ [lorem ipsum foo bar]
```

Note that commas have no special meanings and are valid bareword characters, so
don't use them to separate elements:

```elvish-transcript
~> li = [a, b]
~> put $li
▶ [a, b]
~> put $li[0]
▶ a,
```

## Map

Maps are also surrounded by square brackets; a key/value pair is written
`&key=value` (reminiscent to HTTP query parameters), and pairs are separated by
whitespaces. Whitespaces are allowed after `=`, but not before `=`. They are
one of the basic container types in Elvish. Examples:

```elvish-transcript
~> put [&foo=bar &lorem=ipsum]
▶ [&foo=bar &lorem=ipsum]
~> put [&a=   10
        &b=   23
        &sum= (+ 10 23)]
▶ [&a=10 &b=23 &sum=33]
```

An empty map is written as `[&]`.

If you only specify a key without `=` or a value that follows it, the value will
be `$true`. However, if you keep `=` but don't specify any value after it, the
value will be an empty string. Example:

```elvish-transcript
~> echo [&a &b=]
[&a=$true &b='']
```

# Variable

Variables are named holders of values. The following characters can be used in
variable names (a subset of bareword characters):

-   ASCII letters (a-z and A-Z) and numbers (0-9);

-   The symbols `-_:~`. The colon `:` is special; it is normally used for
    separating namespaces or denoting namespace variables;

-   Non-ASCII codepoints that are printable, as defined by
    [unicode.IsPrint](https://godoc.org/unicode#IsPrint) in Go's standard
    library.

In most other shells, variables can map directly to environmental variables:
`$PATH` is the same as the `PATH` environment variable. This is not the case in
Elvish. Instead, environment variables are put in a dedicated `E:` namespace;
the environment variable `PATH` is known as `$E:PATH`. The `$PATH` variable, on
the other hand, does not exist initially, and if you have defined it, only lives
in a certain lexical scope within the Elvish interpreter.

You will notice that variables sometimes have a leading dollar `$`, and
sometimes not. The tradition is that they do when they are used for their
values, and do not otherwise (e.g. in assignment). This is consistent with most
other shells.

## Assignment

A variable can be assigned by writing its name, `=`, and the value to assign.
There must be inline whitespaces both before and after `=`. Example:

```elvish-transcript
~> foo = bar
```

You can assign multiple values to multiple variables simultaneously, simply by
writing several variable names (separated by inline whitespaces) on the
left-hand side, and several values on the right-hand side:

```elvish-transcript
~> x y = 3 4
```

## Referencing

Use a variable by adding `$` before the name:

```elvish-transcript
~> foo = bar
~> x y = 3 4
~> put $foo
▶ bar
~> put $x
▶ 3
```

Variables must be assigned before use. Attempting to use an unassigned variable
causes a compilation error:

```elvish-transcript
~> echo $x
Compilation error: variable $x not found
[tty], line 1: echo $x
~> { echo $x }
Compilation error: variable $x not found
[tty], line 1: { echo $x }
```

## Explosion and Rest Variable

If a variable contains a list value, you can add `@` before the variable name to
get all its element values. This is called **exploding** the variable:

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

(This notation is restricted to exploding variables. To explode arbitrary
values, use the builtin [explode](builtin.html#explode) command.)

When assigning variables, if you prefix the name of the last variable with `@`,
it gets assigned a list containing all remaining values. That variable is called
a **rest variable**. Example:

```elvish-transcript
~> a b @rest = 1 2 3 4 5 6 7
~> put $a $b $rest
▶ 1
▶ 2
▶ [3 4 5 6 7]
```

Schematically this is a reverse operation to variable explosion, which is why
they share the `@` sign.

## Temporary Assignment

You can prepend a command with **temporary assignments**, which gives variables
temporarily values during the execution of that command.

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

Since ordinary assignments are also a kind of command, they can also be
prepended with temporary assignments:

```elvish-transcript
~> x=1
~> x=100 y = (+ 133 $x)
~> put $x $y
▶ 1
▶ 233
```

Temporary assignments must all appear before the command. As soon as something
that is not a temporary assignments is parsed, Elvish no longer parses temporary
assignments. For instance, in `x=1 echo x=1`, the second `x=1` is not a
temporary assignment, but a bareword.

**Note**: Elvish's behavior differs from bash (or zsh) in one important place.
In bash, temporary assignments to variables do not affect their direct
appearance in the command:

```sh-transcript
bash-4.4$ x=1
bash-4.4$ x=100 echo $x
1
```

## Scoping rule

Elvish has lexical scoping. Scopes are introduced by [lambdas](#lambda) or
[user-defined modules](#user-defined-modules).

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
the `builtin:` namespace, and if that also fails, cause an error:

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
cannot be found, it will be created in the current scope:

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

# Lambda

A function literal, or lambda, is a [code chunk](#code-chunk) surrounded by
curly braces:

```elvish-transcript
~> f = { echo "Inside a lambda" }
~> put $f
▶ <closure 0x18a1a340>
```

One or more whitespace characters after `{` is required: Elvish relies on the
presence of whitespace to disambiguate lambda literals and
[braced lists](#braced-lists). It is good style to put some whitespace before
the closing `}` as well, but this is not required by the syntax.

Functions are first-class values in Elvish. They can be kept in variables, used
as arguments, output on the value channel, and embedded in other data
structures. They can also be used as commands:

```elvish-transcript
~> $f
Inside a lambda
~> { echo "Inside a literal lambda" }
Inside a literal lambda
```

The last command resembles a code block in C-like languages in syntax. But under
the hood, it defines a function on the fly and calls it immediately.

Functions defined using the basic syntax above do not accept any arguments or
options. To do so, you need to write a signature.

## Signature

A **signature** specifies the arguments a function can accept:

```elvish-transcript
~> f = [a b]{ put $b $a }
~> $f lorem ipsum
▶ ipsum
▶ lorem
```

There should be no space between `]` and `{`; otherwise Elvish will parse the
signature as a list, followed by a lambda without signature:

```elvish-transcript
~> put [a]{ nop }
▶ <closure 0xc420153d80>
~> put [a] { nop }
▶ [a]
▶ <closure 0xc42004a480>
```

Like in the left hand of assignments, if you prefix the last argument with `@`,
it becomes a **rest argument**, and its value is a list containing all the
remaining arguments:

```elvish-transcript
~> f = [a @rest]{ put $a $rest }
~> $f lorem
▶ lorem
▶ []
~> $f lorem ipsum dolar sit
▶ lorem
▶ [ipsum dolar sit]
```

You can also declare options in the signature. The syntax is `&name=default`
(like a map pair), where `default` is the default value for the option:

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

## Closure Semantics

User-defined functions are also known as "closures", because they have
[closure semantics](<https://en.wikipedia.org/wiki/Closure_(computer_programming)>).

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

# Indexing

Indexing is done by putting one or more **index expressions** in brackets `[]`
after a value.

## List Indexing

Lists can be indexed with any of the following:

-   A non-negative integer, an offset counting from the beginning of the list.
    For example, `$li[0]` is the first element of `$li`.

-   A negative integer, an offset counting from the back of the list. For
    instance, `$li[-1]` is the last element `$li`.

-   A slice `$a:$b`, where both `$a` and `$b` are integers. The result is
    sublist of `$li[$a]` up to, but not including, `$li[$b]`. For instance,
    `$li[4:7]` equals `[$li[4] $li[5] $li[6]]`, while `$li[1:-1]` contains all
    elements from `$li` except the first and last one.

    Both integers may be omitted; `$a` defaults to 0 while `$b` defaults to the
    length of the list. For instance, `$li[:2]` is equivalent to `$li[0:2]`,
    `$li[2:]` is equivalent to `$li[2:(count $li)]`, and `$li[:]` makes a copy
    of `$li`. The last form is rarely useful, as lists are immutable.

    Note that the slice needs to be a **single** string, so there cannot be any
    spaces within the slice. For instance, `$li[2:10]` cannot be written as
    `$li[2: 10]`; the latter contains two indicies and is equivalent to
    `$li[2:] $li[10]` (see [Multiple Indicies](#multiple-indicies)).

-   Not yet implemented: The string `@`. The result is all the values in the
    list. Note that this is not the same as `:`: if `$li` has 10 elements,
    `$li[@]` evaluates to 10 values (all the elements in the list), while
    `$li[:]` evaluates to just one value (a copy of the list).

    When used on a variable like `$li`, it is equivalent to the explosion
    construct `$li[@]`. It is useful, however, when used on other constructs,
    like output capture or other

Examples:

```elvish-transcript
~> li = [lorem ipsum foo bar]
~> put $li[0]
▶ lorem
~> put $li[-1]
▶ bar
~> put $li[0:2]
▶ [lorem ipsum]
```

(Negative indicies and slicing are borrowed from Python.)

## String indexing

**NOTE**: String indexing will likely change.

Strings should always be UTF-8, and they can indexed by **byte indicies at which
codepoints start**, and indexing results in **the codepoint that starts there**.
This is best explained with examples:

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
    5\. Hence valid indicies are 0 and 3:

    ```elvish-transcript
    ~> put 世界[0]
    ▶ 世
    ~> put 世界[3]
    ▶ 界
    ```

Strings can also be indexed by slices.

(This idea of indexing codepoints by their byte positions is borrowed from
Julia.)

## Map indexing

Maps are simply indexed by their keys. There is no slice indexing, and `:` does
not have a special meaning. Examples:

```elvish-transcript
~> map = [&a=lorem &b=ipsum &a:b=haha]
~> echo $map[a]
lorem
~> echo $map[a:b]
haha
```

## Multiple Indices

If you put multiple values in the index, you get multiple values: `$li[x y z]`
is equivalent to `$li[x] $li[y] $li[z]`. This applies to all indexable values.
Examples:

```elvish-transcript
~> put elv[0 2 0:2]
▶ e
▶ v
▶ el
~> put [lorem ipsum foo bar][0 2 0:2]
▶ lorem
▶ foo
▶ [lorem ipsum]
~> put [&a=lorem &b=ipsum &a:b=haha][a a:b]
▶ lorem
▶ haha
```

# Output Capture

Output capture is formed by putting parentheses `()` around a
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

# Exception Capture

Exception capture is formed by putting `?()` around a code chunk. It runs the
chunk and evaluates to the exception it throws.

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

Exception captures do not affect the output of the code chunk. You can combine
output capture and exception capture:

```elvish
output = (error = ?(commands-that-may-fail))
```

# Tilde Expansion

Tildes are special when they appear at the beginning of an expression (the exact
meaning of "expression" will be explained later). The string after it, up to the
first `/` or the end of the word, is taken as a user name; and they together
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

If you need them to be, surround them with braces (the reason this works will be
explained later):

```elvish-transcript
~> put a{~root}
▶ a/root
```

# Wildcard Patterns

**Wildcard patterns** are patterns containing **wildcards**, and they evaluate
to all filenames they match.

We will use this directory tree in examples:

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

## Modifiers

Wildcards can be **modified** using the same syntax as indexing. For instance,
in `*[match-hidden]` the `*` wildcard is modified with the `match-hidden`
modifier. Multiple matchers can be chained like `*[set:abc][range:0-9]`. There
are two kinds of modifiers:

**Global modifiers** apply to the whole pattern and can be placed after any
wildcard:

-   `nomatch-ok` tells Elvish not to throw an error when there is no match for
    the pattern. For instance, in the example directory `put bad*` will be an
    error, but `put bad*[nomatch-ok]` does exactly nothing.

-   `but:xxx` (where `xxx` is any filename) excludes the filename from the final
    result.

Although global modifiers affect the entire wildcard pattern, you can add it
after any wildcard, and the effect is the same. For example,
`put */*[nomatch-ok].cpp` and `put *[nomatch-ok]/*.cpp` do the same thing.

On the other hand, you must add it after a wildcard, instead of after the entire
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

    -   Multiple matchers, they are OR'ed. For instance, ?[set:aeoiu][digit]
        matches `aeoiu` and digits.

    -   Dots at the beginning of filenames always require an explicit
        `match-hidden`, even if the matcher includes `.`. For example,
        `?[set:.a]x.conf` does **not** match `.x.conf`; you have to
        `?[set:.a match-hidden]x.conf`.

    -   Likewise, you always need to use `**` to match slashes, even if the
        matcher includes `/`. For example `*[set:abc/]` is the same as
        `*[set:abc]`.

# Compound Expression and Braced Lists

Writing several expressions together with no space in between will concatenate
them. This creates a **compound expression**, because it mimics the formation of
compound words in natural languages. Examples:

```elvish-transcript
~> put 'a'b"c" # compounding three string literals
▶ abc
~> v = value
~> put '$v is '$v # compounding one string literal with one string variable
▶ '$v is value'
```

Many constructs in Elvish can generate multiple values, like indexing with
multiple indices and output captures. Compounding multiple values with other
values generates all possible combinations:

```elvish-transcript
~> put (put a b)-(put 1 2)
▶ a-1
▶ a-2
▶ b-1
▶ b-2
```

Note the order of the generated values. The value that comes later changes
faster.

**NOTE**: There is a perhaps a better way to explain the ordering, but you can
think of the previous code as equivalent to this:

```elvish-transcript
for x [a b] {
  for y [1 2] {
    put $x-$y
  }
}
```

## Braced Lists

In practice, you never have to write `(put a b)`: you can use a braced list
`{a,b}`:

```elvish-transcript
~> put {a,b}-{1,2}
▶ a-1
▶ a-2
▶ b-1
▶ b-2
```

Elements in braced lists can also be separated with whitespaces, or a
combination of comma and whitespaces (the latter not recommended):

```elvish-transcript
~> put {a b , c,d}
▶ a
▶ b
▶ c
▶ d
```

(In future, the syntax might be made more strict.)

Braced list is merely a syntax for grouping multiple values. It is not a data
structure.

# Expression Structure and Precedence

Braced lists are evaluated before being compounded with other values. You can
use this to affect the order of evaluation. For instance, `put *.txt` gives you
all filenames that end with `.txt` in the current directory; while `put {*}.txt`
gives you all filenames in the current directory, appended with `.txt`.

**TODO**: document evaluation order regarding tilde and wildcards.

# Ordinary Command

The **command** is probably the most important syntax construct in shell
languages, and Elvish is no exception. The word **command** itself, is
overloaded with meanings. In the terminology of this document, the term
**command** include the following:

-   An ordinary assignment, introduced above;

-   An ordinary command, introduced in this section;

-   A special command, introduced in the next section.

An **ordinary command** consists of a compulsory head, and any number of
arguments, options and redirections.

## Head

The **head** must appear first. It is an arbitrary word that determines what
will be run. Examples:

```elvish-transcript
~> ls -l # the string ls is the head
(output omitted)
~> (put [@a]{ ls $@a }) -l
(same output)
```

The head must evaluate to one value. For instance, the following does not work:

```elvish-transcript
~> (put [@a]{ ls $@a } -l)
Exception: head of command must be a single value; got 2 values
[tty], line 1: (put [@a]{ ls $@a } -l)
```

The definition of barewords is relaxed for the head to include `<`, `>`, `*` and
`^`. These are all names of numeric builtins:

```elvish-transcript
~> < 3 5 # less-than
▶ $true
~> > 3 5 # greater-than
▶ $false
~> * 3 5 # multiplication
▶ 15
~> ^ 3 5 # power
▶ 243
```

## Arguments and Options

**Arguments** (args for short) and **options** (opts for short) can be supplied
to commands. Arguments are arbitrary words, while options have the same syntax
as map pairs. They are separated by inline whitespaces:

```elvish-transcript
~> echo &sep=, a b c # seq=, is an option; a b c are arguments
a,b,c
```

Like in maps, `&key` is equivalent to `&key=$true`:

```elvish-transcript
~> fn f [&opt=$false]{ put $opt }
~> f &opt
▶ $true
```

## Redirections

Redirections are used for modifying file descriptors (FD).

The most common form of redirections opens a file and associates it with an FD.
The form consists of an optional destination FD (like `2`), a redirection
operator (like `>`) and a filename (like `error.log`):

-   The **destination fd** determines which FD to modify. It can be given either
    as a number, or one of `stdin`, `stdout` and `stderr`. There must be no
    space between the FD and the redirection operator; otherwise Elvish will
    parse it as an argument.

    The destination FD can be omitted, in which case it is inferred from the
    redirection operator.

-   The **redirection operator** determines the mode to open the file, and the
    destination FD if it is not explicitly specified.

-   The **filename** names the file to open.

Possible redirection operators and their default FDs are:

-   `<` for reading. The default FD is 0 (stdin).

-   `>` for writing. The default FD is 1 (stdout).

-   `>>` for appending. The default FD is 1 (stdout).

-   `<>` for reading and writing. The default FD is 1 (stdout).

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

Redirections can also be used for closing or duplicating FDs. Instead of writing
a filename, use `&fd` (where `fd` is a number, or any of `stdin`, `stdout` and
`stderr`) for duplicating, or `&-` for closing. In this case, the redirection
operator only determines the default destination FD (and is totally irrevelant
if a destination FD is specified). Examples:

```elvish-transcript
~> ls >&- # close stdout
/bin/ls: write error: Bad file descriptor
Exception: ls exited with 2
Traceback:
  [interactive], line 1:
    ls >&-
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

## Ordering

Elvish does not impose any ordering of arguments, options and redirections: they
can intermix each other. The only requirement is that the head must come first.
This is different from POSIX shells, where redirections may appear before the
head. For instance, the following two both work in POSIX shell, but only the
former works in Elvish:

```sh
echo something > file
> file echo something # mistaken for the comparison builtin ">" in Elvish
```

# Special Commands

**Special commands** obey the same syntax rules as normal commands (i.e.
syntactically special commands can be treated the same as ordinary commands),
but have evaluation rules that are custom to each command. To explain this, we
use the following example:

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

## Deleting variable or element: `del`

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

## Logics: `and` and `or`

The `and` special command evaluates its arguments from left to right; as soon as
a booleanly false value is obtained, it outputs the value and stops. When given
no arguments, it outputs `$true`.

The `or` special command is the same except that it stops when a booleanly true
value is obtained. When given no arguments, it outpus `$false`.

## Condition: `if`

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

## Conditional Loop: `while`

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

## Iterative Loop: `for`

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

## Exception Control: `try`

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

## Function Definition: `fn`

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

Under the hood, `fn` defines a variable with the given name plus `~` (see
[command resolution](#command-resolution)).

# Command Resolution

When using a literal string as the head of a command, it is first **resolved**
during the compilation phase, using the following order:

1.  If the name matches any of the [special commands](#special-commands), it is
    treated as so.

2.  Finding a variable with the name of the command plus a `~` suffix.

    For instance, given a command `f a b`, Elvish looks for the variable `$f~`,
    using the ordinary variable [scoping rule](#scoping-rule), except that
    resolution failures do not cause errors but fall back to the next step.

    Functions defined with `fn` as well as builtin functions are actually
    variables with a `~` suffix in their names.

3.  External commands.

    This step always succeeds during compilation, even if the command does not
    exist. Later, during evaluation, a **searching** step determines whether the
    external command exists.

The entire resolution procedure can be emulated with the
[resolve](builtin.html#resolve) command. Searching of external commands can be
emulated with the [search-external](builtin.html#search-builtin) command.

**TIP**: Step 2 of the command resolution rules means that if you define a variable
with a name ending with `~`, you can use it as a command:

```elvish-transcript
~> f~ = { put f }
~> f
▶ f
```

The same also applies to function parameters:

```elvish-transcript
~> fn g [f~]{ f }
~> g { put f }
▶ f
```

# Pipeline

A **pipeline** is formed by joining one or more commands together with the pipe
sign (`|`).

## IO Semantics

For each pair of adjacent commands `a | b`, the output of `a` is connected to
the input of `b`. Both the byte pipe and the value channel are connected, even
if one of them is not used.

Command redirections are applied before the connection happens. For instance,
the following writes `foo` to `a.txt` instead of the output:

```elvish-transcript
~> echo foo > a.txt | cat
~> cat a.txt
foo
```

## Execution Flow

All of the commands in a pipeline are executed in parallel, and the execution of
the pipeline finishes when all of its commands finish execution.

If one or more command in a pipeline throws an exception, the other commands
will continue to execute as normal. After all commands finish execution, an
exception is thrown, the value of which depends on the number of commands that
have thrown an exception:

-   If only one command has thrown an exception, that exception is rethrown.

-   If more than one commands have thrown exceptions, a "composite exception",
    containing information all exceptions involved, is thrown.

## Background Pipeline

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
commands that raise special "flow control" exceptions. The `for` and `while`
commands capture `break` and `continue`, while `fn` modifies its closure to
capture `return`.

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

**WARNING:** The exception data type currently supports a single attribute,
`cause`, that can be used to extract an object describing the cause
of the exception; e.g. `$e[cause]`.  This is not a string. This is an
experimental feature. You should probably use `(to-string $e)` at this
time in any production code.

# Namespaces and Modules

Namespace in Elvish helps prevent name collisions and is important for building
modules.

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

## Special Namespaces

The following namespaces have special meanings to the language:

-   `local:` and `up:` refer to lexical scopes, and have been documented above.

-   `e:` refers to externals. For instance, `e:ls` refers to the external
    command `ls`.

    Most of the time you can rely on the rules of
    [command resolution](#command-resolution) and do not need to use this
    explicitly, unless a function defined by you (or an Elvish builtin) shadows
    an external command.

-   `E:` refers to environment variables. For instance, `$E:USER` is the
    environment variable `USER`.

    This **is** always needed, because unlike command resolution, variable
    resolution does not fall back onto environment variables.

-   `builtin:` refers to builtin functions and variables.

    You don't need to use this explicitly unless you have defined names that
    shadows builtin counterparts.

## Pre-Defined Modules

Namespaces that are not special (i,e. one of the above) are also called
**modules**. Aside from these special namespaces, Elvish also comes with the
following modules:

-   `edit:` for accessing the Elvish editor. This module is available in
    interactive mode and does not need importing.

    See [reference](edit.html).

-   `re:` for regular expression facilities. This module is always available.
    See [reference](re.html).

-   `daemon:` for manipulating the daemon. This module is always available.

    This is not yet documented.

## User-Defined Modules

You can define your own modules with Elvishscript by putting them under
`~/.elvish/lib` and giving them a `.elv` extension. For instance, to define a
module named `a`, store it in `~/.elvish/lib/a.elv`:

```elvish-transcript
~> cat ~/.elvish/lib/a.elv
echo "mod a loading"
fn f {
  echo "f from mod a"
}
```

To import the module, use `use`:

```elvish-transcript
~> use a
mod a loading
~> a:f
f from mod a
```

Modules in nested directories can also be imported. For example, if you have
defined a module in `~/.elvish/lib/x/y/z.elv`, you can import it by using `use
x/y/z`, and the resulting namespace will be `z:`:

```elvish-transcript
~> cat .elvish/lib/x/y/z.elv
fn f {
  echo "f from x/y/z"
}
~> use x/y/z
~> z:f
f from x/y/z
```

In general, if you import a module from a nested directory, the resulting
namespace will be the same as the file name (without the `.elv` extension).

## Aliasing

You can import a module as a namespace of your choice by specifying a second
argument to `use`. For example, to import `x/y/z` as a `xyz` namespace, you can
use `use x/y/z xyz`:

```elvish-transcript
~> use x/y/z xyz
~> xyz:f
f from x/y/z
```

This is especially useful when you need to import several modules that are in
different directories but have the same file name.

## Scoping of Imports

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

## Re-Importing

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
use a/b
```
