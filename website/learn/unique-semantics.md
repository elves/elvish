<!-- toc -->

The semantics of Elvish is unique in many aspects when compared to other shells.
This can be surprising if you are used to other shells, and it is a result of
the design choice of making Elvish a full-fledged programming language.

# Structureful IO

Elvish offers the ability to build elaborate data structures, "return" them from
functions, and pass them through pipelines.

## Motivation

Traditional shells use strings for all kinds of data. They can be stored in
variables, used as function arguments, written to output and read from input.
Strings are very simple to use, but they fall short if your data has an inherent
structure. A common solution is using pseudo-structures of "each line
representing a record, and each (whitespace-separated) field represents a
property", which is fine as long as your data do not contain whitespaces. If
they do, you will quickly run into problems with escaping and quotation and find
yourself doing black magics with strings.

Some shells provide data structures like lists and maps, but they are usually
not first-class values. You can store them in variables, but you might not to be
able to nest them, pass them to functions, or returned them from functions.

## Data Structures and "Returning" Them

Elvish offers first-class support for data structures such as lists and maps.
Here is an example that uses a list:

```elvish-transcript
~> var li = [foo bar 'lorem ipsum']
~> kind-of $li # "kind" is like type
▶ list
~> count $li # count the number of elements in a list
▶ 3
```

(See the [language reference](../ref/language.html) for a more complete
description of the builtin data structures.)

As you can see, you can store lists in variables and use them as command
arguments. But they would be much less useful if you cannot **return** them from
a function. A naive way to do this is by `echo`ing the list and use output
capture to recover it:

```elvish-transcript
~> fn f {
     echo [foo bar 'lorem ipsum']
   }
~> var li = (f) # (...) is output capture, like $(...) in other shells
~> kind-of $li
▶ string
~> count $li # count the number of bytes, since $li is now a string
▶ 23
```

As we have seen, our attempt to output a list has turned it into a string. This
is because the `echo` command in Elvish, like in other shells, is
string-oriented. To echo a list, it has to be first converted to a string.

Elvish provides a `put` command to output structured values as they are:

```elvish-transcript
~> fn f {
     put [foo bar 'lorem ipsum']
   }
~> var li = (f)
~> kind-of $li
▶ list
~> count $li
▶ 3
```

So how does `put` work differently from `echo` under the hood?

In Elvish, the standard output is made up of two parts: one traditional
byte-oriented **file**, and one internal **value-oriented channel**. The `echo`
command writes to the file, so it has to serialize its arguments into strings;
the `put` command writes to the value-oriented channel, preserving all the
internal structures of the values.

If you invoke `put` directly from the command prompt, the values it output have
a leading `▶`:

```elvish-transcript
~> put [foo bar]
▶ [foo bar]
```

The leading arrow is a way to visualize that a command has written something
onto the value channel, and not part of the value itself.

In retrospect, you may discover that the `kind-of` and `count` builtin commands
also write their output to the value channel.

## Passing Data Structures Through the Pipeline

When I said that standard output in Elvish comprises two parts, I was not
telling the full story: pipelines in Elvish also have these two parts, in a very
similar way. Data structures can flow in the value-oriented part of the pipeline
as well. For instance, the `each` command takes **input** from the
value-oriented channel, and apply a function to each value:

```elvish-transcript
~> put lorem ipsum | each {|x| echo "Got "$x }
Got lorem
Got ipsum
```

There are many builtin commands that inputs or outputs values. As another
example, the `take` commands retains a fixed number of items:

```elvish-transcript
~> put [lorem ipsum] "foo\nbar" [&key=value] | take 2
▶ [lorem ipsum]
▶ "foo\nbar"
```

## Interoperability with External Commands

Unfortunately, the ability of passing structured values is not available to
external commands. However, Elvish comes with a pair of commands for JSON
serialization/deserialization. The following snippet illustrates how to
interoperate with a Python script:

```elvish-transcript
~> cat sort-list.py
import json, sys
li = json.load(sys.stdin)
li.sort()
json.dump(li, sys.stdout)
~> put [lorem ipsum foo bar] | to-json | python sort-list.py | from-json
▶ [bar foo ipsum lorem]
```

It is easy to write a wrapper for such external commands:

```elvish-transcript
~> fn sort-list { to-json | python sort-list.py | from-json }
~> put [lorem ipsum foo bar] | sort-list
▶ [bar foo ipsum lorem]
```

More serialization/deserialization commands may be added to the language in the
future.

# Exit Status and Exceptions

Unix commands exit with a non-zero value to signal errors. This is available
traditionally as a `$?` variable in other shells:

```bash
true
echo $? # prints "0"
false
echo $? # prints "1"
```

Builtin commands and user-defined functions also do this to signal errors,
although they are not Unix commands:

```bash
bad() {
  return 2
}
bad
echo $? # prints "2"
```

This model is fine, only if most errors are non-fatal (so that errors from a
previous command normally do not affect the execution of subsequence ones) and
the script author remembers to check `$?` for the rare fatal errors.

Elvish has no concept of exit status. Instead, it has exceptions that, when
thrown, interrupt the flow of execution. The equivalency of the `bad` function
in elvish is as follows:

```elvish
fn bad {
  fail "bad things have happened" # throw an exception
}
bad # will print a stack trace and stop execution
echo "after bad" # not executed
```

(If you run this interactively, you need to enter a literal newline after `bad`
by pressing <kbd>Alt-Enter</kbd> to make sure that it is executed in the same
chunk as `echo "after bad"`.)

And, non-zero exit status from external commands are turned into exceptions:

```elvish
false # will print a stack trace and stop execution
echo "after false"
```

An alternative way to describe this is that Elvish **does** have exit statuses,
but non-zero exit statuses terminates execution by default. You can handle
non-zero exit statuses by wrapping the command in a
[`try`](../ref/language.html#try) block.

Compare with POSIX shells, the behavior of Elvish is similar to `set -e` or
`set -o errexit`, or having implicit `&&` operators joining all the commands.
Defaulting to stopping execution when bad things happen makes Elvish safer and
code behavior more predictable.

## Predicates and `if`

The use of exit status is not limited to errors, however. In the Unix toolbox,
quite a few commands exit with 0 to signal "true" and 1 to signal "false".
Notably ones are:

-   `test` aka `[`: testing file types, comparing numbers and strings;

-   `grep`: exits with 0 when there are matches, with 1 otherwise;

-   `diff`: exits with 0 when files are the same, with 1 otherwise;

-   `true` and `false`, always exit with 0 and 1 respectively.

The `if` control structure in POSIX shell is designed to work with such
predicate commands: it takes a pipeline, and executes the body if the last
command in the pipeline exits with 0. Examples:

```sh
# command 1
if true; then
  echo 'always executes'
fi

# command 2
n=10
if test $n -gt 2; then
  echo 'executes when $n > 2'
fi

# command 3
if diff a.txt b.txt; then
  echo 'a.txt and b.txt are the same'
fi
```

Since Elvish treats non-zero exit status as a kind of exception, the way that
predicate commands and `if` work in POSIX shell does not work well for Elvish.
Instead, Elvish's `if` is like most non-shell programming languages: it takes a
value, and executes the body if the value is booleanly true. The first command
above is written in Elvish as:

```elvish
if $true {
  echo 'always executes'
}
```

The way to write the second command in Elvish warrants an explanation of how
predicates work in Elvish first. Predicates in Elvish simply write a boolean
output, either `$true` or `$false`:

```elvish-transcript
~> > 10 2
▶ $true
~> > 1 2
▶ $false
```

To use predicates in `if`, you simply capture its output with `()`. So the
second command is written in Elvish as:

```elvish
var n = 10
if (> $n 2) {
  echo 'executes when $n > 2'
}
```

The parentheses after `if` are different to those in C: In C it is a syntactical
requirement to put them around the condition; in Elvish, it functions as output
capture operator.

Sometimes it can be useful to have a condition on whether an external commands
exits with 0. In this case, you can use the exception capture operator `?()`:

```elvish
if ?(diff a.txt b.txt) {
  echo 'a.txt and b.txt are the same'
}
```

In Elvish, all exceptions are booleanly false, while the special `$ok` value is
booleanly true. If the `diff` exits with 0, the `?(...)` construct evaluates to
`$ok`, which is booleanly true. Otherwise, it evaluates to an exception, which
is booleanly false. Overall, this leads to a similar semantics with the POSIX
`if` command.

Note that the following code does have a severe downside: `?()` will prevent any
kind of exceptions from throwing. In this case, we only want to turn one sort of
exception into a boolean: `diff` exits with 1. If `diff` exits with 2, it
usually means that there was a genuine error (e.g. `a.txt` does not exist).
Swallowing this error defeats Elvish's philosophy of erring on the side of
caution; a more sophisticated system of handling exit status is still being
considered.

# Phases of Code Execution

A piece of code that gets evaluated as a whole is called a **chunk** (a loanword
from Lua). If you run `elvish some-script.elv` from the command line, the entire
script is one chunk; in interactive mode, each time you hit Enter, the code you
have written is one chunk.

Elvish interprets a code chunk in 3 phases: it first **parse**s the code into a
syntax tree, then **compile**s the syntax tree code to an internal
representation, and finally **evaluate**s the just-generated internal
representation.

If any error happens during the first two phases, Elvish rejects the chunk
without executing any of it. For instance, in Elvish unclosed parenthesis is an
error during the parsing phase. The following code, when executed as a chunk,
does nothing other than printing the parse error:

```elvish
echo before
echo (
```

The same code, interpreted as bash, also contains a syntax error. However, if
you save this file to `bad.bash` and run `bash bad.bash`, bash will execute the
first line before complaining about the syntax error on the second line.

Likewise, in Elvish using an unassigned variable is a compilation error, so the
following code does nothing either:

```elvish
# assuming $nonexistent was not assigned
echo before
echo $nonexistent
```

There seems to be no equivalency of compilation errors in other shells, but this
extra compilation phase makes the language safer. In future, optional type
checking may be introduced, which will fit into the compilation phase.

# Assignment Semantics

In Python, JavaScript and many other languages, if you assign a container (e.g.
a map) to multiple variables, modifications via those variables mutate the same
container. This is best illustrated with an example:

```python
m = {'foo': 'bar', 'lorem': 'ipsum'}
m2 = m
m2['foo'] = 'quux'
print(m['foo']) # prints "quux"
```

This is because in such languages, variables do not hold the "actual" map, but a
reference to it. After the assignment `m2 = m`, both variables refer to the same
map. The subsequent element assignment `m2['foo'] = 'quux'` mutates the
underlying map, so `m['foo']` is also changed.

This is not the case for Elvish:

```elvish-transcript
~> var m = [&foo=bar &lorem=ipsum]
~> var m2 = $m
~> set m2[foo] = quux
~> put $m[foo]
▶ bar
```

It seems that when you assign `m2 = $m`, the entire map is copied from `$m` into
`$m2`, so any subsequent changes to `$m2` does not affect the original map in
`$m`. You can entirely think of it this way: thinking **assignment as copying**
correctly models the behavior of Elvish.

But wouldn't it be expensive to copy an entire list or map every time assignment
happens? No, the "copying" is actually very cheap. Is it implemented as
[copy-on-write](https://en.wikipedia.org/wiki/Copy-on-write) -- i.e. the copying
is delayed until `$m2` gets modified? No, subsequent modifications to the new
`$m2` is also very cheap. Read on if you are interested in how it is possible.

## Implementation Detail: Persistent Data Structures

Like in Python and JavaScript, Elvish variables like `$m` and `$m2` also only
hold a reference to the underlying map. However, that map is **immutable**,
meaning that they never change after creation. That explains why `$m` did not
change: because the map `$m` refers to never changes. But how is it possible to
do `m2[foo] = quux` if the map is immutable?

The map implementation of Elvish has another property: although the map is
immutable, it is easy to create a slight variation of one map. Given a map, it
is easy to create another map that is almost the same, either 1) with one more
key/value pair, or 2) with the value for one key changed, or 3) with one fewer
key/value pair. This operation is fast, even if the original map is very large.

This low-level functionality is exposed by the `assoc` (associate) and `dissoc`
(dissociate) builtins:

```elvish-transcript
~> assoc [&] foo quux # "add" one pair
▶ [&foo=quux]
~> assoc [&foo=bar &lorem=ipsum] foo quux # "modify" one pair
▶ [&lorem=ipsum &foo=quux]
~> dissoc [&foo=bar &lorem=ipsum] foo # "remove" one pair
▶ [&lorem=ipsum]
```

Now, although maps are immutable, variables are mutable. So when you try to
assign an element of `$m2`, Elvish turns that into an assignment of `$m2`
itself:

```elvish
set m2[foo] = quux
# is just syntax sugar for:
set m2 = (assoc $m2 foo quux)
```

The sort of immutable data structures that support cheap creation of "slight
variations" are called
[persistent data structures](https://en.wikipedia.org/wiki/Persistent_data_structure)
and is used in functional programming languages. However, the way Elvish turns
assignment to `$m2[foo]` into an assignment to `$m2` itself seems to be a new
approach.
