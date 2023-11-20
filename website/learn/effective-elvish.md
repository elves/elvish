<!-- toc -->

Elvish is not an entirely new language. Its programming techniques have two
primary sources: traditional Unix shells and functional programming languages,
both dating back to many decades ago. However, the way Elvish combines those two
paradigms is unique in many ways, which enables new ways to write code.

This document is an advanced tutorial focusing on how to write idiomatic Elvish
code, code that is concise and clear, and takes full advantage of Elvish's
features.

An appropriate adjective for idiomatic Elvish code, like *Pythonic* for Python
or *Rubyesque* for Ruby, is **Elven**. In
[Roguelike games](https://en.wikipedia.org/wiki/Roguelike), Elven items are
known to be high-quality, artful and resilient. So is Elven code.

# Style

## Naming

Use `dash-delimited-words` for names of variables and functions. Underscores are
allowed in variable and function names, but their use should be limited to
environment variables (e.g. `$E:LC_ALL`) and external commands (e.g. `pkg_add`).

When building a module, use a leading dash to communicate that a variable or
function is subject to change in future and cannot be relied upon, either
because it is an experimental feature or implementation detail.

Elvish's core libraries follow the naming convention above.

## Indentation

Indent by two spaces.

## Code Blocks

In Elvish, code blocks in control structures are delimited by curly braces. This
is perhaps the most visible difference of Elvish from most other shells like
bash, zsh or fish. The following bash code:

```bash
if true; then
  echo true
fi
```

Is written like this in Elvish:

```elvish
if $true {
  echo true
}
```

If you have used lambdas in Elvish, you will notice that code blocks are
syntactically just parameter-list-less lambdas.

In Elvish, you cannot put opening braces of code blocks on the next line. This
won't work:

```elvish
if $true
{ # wrong!
  echo true
}
```

Instead, you must write:

```elvish
if $true {
  echo true
}
```

This is because in Elvish, control structures like `if` follow the same syntax
as normal commands, hence newlines terminate them. To make the code block part
of the `if` command, it must appear on the same line.

# Using the Pipeline

Elvish is equipped with a powerful tool for passing data: the pipeline. Like in
traditional shells, it is an intuitive notation for data processing: data flows
from left to right, undergoing one transformation after another. Unlike in
traditional shells, it is not restricted to unstructured bytes: all Elvish
values, including lists, maps and even closures, can flow in the pipeline. This
section documents how to make the most use of pipelines.

## Returning Values with Structured Output

Unlike functions in most other programming languages, Elvish commands do not
have return values. Instead, they can write to *structured output*, which is
similar to the traditional byte-based stdout, but preserves all internal
structures of arbitrary Elvish values. The most fundamental command that does
this is `put`:

```elvish-transcript
~> put foo
▶ foo
~> var x = (put foo)
~> put $x
▶ foo
```

This is hardly impressive - you can output and recover simple strings using good
old byte-based output as well. But let's try this:

```elvish-transcript
~> put "a\nb" [foo bar]
▶ "a\nb"
▶ [foo bar]
~> var s li = (put "a\nb" [foo bar])
~> put $s
▶ "a\nb"
~> put $li[0]
▶ foo
```

Here, two things are worth mentioning: the first value we `put` contains a
newline, and the second value is a list. When we capture the output, we get
those exact values back. Passing structured data is difficult with byte-based
output, but trivial with value output.

Besides `put`, many other builtin commands and commands in builtin modules also
write to structured output, like `str:split`:

```elvish-transcript
~> use str
~> str:split , foo,bar
▶ foo
▶ bar
~> var words = [(str:split , foo,bar)]
~> put $words
▶ [foo bar]
```

User-defined functions behave in the same way: they "return" values by writing
to structured stdout. Without realizing that "return values" are just outputs in
Elvish, it is easy to think of `put` as **the** command to "return" values and
write code like this:

```elvish-transcript
~> fn split-by-comma {|s| use str; put (str:split , $s) }
~> split-by-comma foo,bar
▶ foo
▶ bar
```

The `split-by-comma` function works, but it can be written more concisely as:

```elvish-transcript
~> fn split-by-comma {|s| use str; str:split , $s }
~> split-by-comma foo,bar
▶ foo
▶ bar
```

In fact, the pattern `put (some-cmd)` is almost always redundant and equivalent
to just `some-command`.

Similarly, it is seldom necessary to write `echo (some-cmd)`: it is almost
always equivalent to just `some-cmd`. As an exercise, try simplifying the
following function:

```elvish
fn git-describe { echo (git describe --tags --always) }
```

## Mixing Bytes and Values

Each pipe in Elvish comprises two components: one traditional byte pipe that
carries unstructured bytes, and one value pipe that carries Elvish values. You
can write to both, and output capture will capture both:

```elvish-transcript
~> fn f { echo bytes; put value }
~> f
bytes
▶ value
~> var outs = [(f)]
~> put $outs
▶ [bytes value]
```

This also illustrates that the output capture operator `(...)` works with both
byte and value outputs, and it can recover the output sent to `echo`. When byte
output contains multiple lines, each line becomes one value:

```elvish-transcript
~> var x = [(echo "lorem\nipsum")]
~> put $x
▶ [lorem ipsum]
```

Most Elvish builtin functions also work with both byte and value inputs.
Similarly to output capture, they split their byte input by newlines. For
example:

```elvish-transcript
~> use str
~> put lorem ipsum | each $str:to-upper~
▶ LOREM
▶ IPSUM
~> echo "lorem\nipsum" | each $str:to-upper~
▶ LOREM
▶ IPSUM
```

This line-oriented processing of byte input is consistent with traditional Unix
tools like `grep`, `sed` and `awk`. In fact, it is easy to write your own `grep`
in Elvish:

```elvish-transcript
~> use re
~> fn mygrep {|p| each {|line| if (re:match $p $line) { echo $line } } }
~> cat in.txt
abc
123
lorem
456
~> cat in.txt | mygrep '[0-9]'
123
456
```

Note that it is more concise to write `mygrep ... < in.txt`.

However, this line-oriented behavior is not always desirable: not all Unix
commands output newline-separated data. When you want to get the output as is,
as a single string, you can use the `slurp` command:

```elvish-transcript
~> echo "a\nb\nc" | slurp
▶ "a\nb\nc\n"
```

One immediate use of `slurp` is to read a whole file into a string:

```elvish-transcript
~> cat hello.go
package main

import "fmt"

func main() {
            fmt.Println("vim-go")
}
~> hello-go = (slurp < hello.go)
~> put $hello-go
▶ "package main\n\nimport \"fmt\"\n\nfunc main()
{\n\tfmt.Println(\"vim-go\")\n}\n"
```

It is also useful, for example, when working with NUL-separated output:

```elvish-transcript
~> touch "a\nb.go"
~> mkdir d
~> touch d/f.go
~> use str
~> find . -name '*.go' -print0 | str:split "\000" (slurp)
▶ "./a\nb.go"
▶ ./d/f.go
▶ ''
```

In the above command, `slurp` turns the input into one string, which is then
used as an argument to `str:split`. The `str:split` command then splits the
whole input by NUL bytes.

Note that in Elvish, strings can contain NUL bytes; in fact, they can contain
any byte; this makes Elvish suitable for working with binary data. (Also, note
that the `find` command terminates its output with a NUL byte, hence we see a
trailing empty string in the output.)

One side note: In the first example, we saw that `bytes` appeared before
`value`. This is not guaranteed: byte output and value output are separate, it
is possible to get `value` before `bytes` in more complex cases. Writes to one
component, however, always have their orders preserved, so in `put x; put y`,
`x` will always appear before `y`.

## Prefer Pipes Over Parentheses

If you have experience with Lisp, you will discover that you can write Elvish
code very similar to Lisp. For instance, to split a string containing
comma-separated value, reduplicate each value (using commas as separators), and
rejoin them with semicolons, you can write:

```elvish-transcript
~> var csv = a,b,foo,bar
~> use str
~> str:join ';' [(each {|x| put $x,$x } [(str:split , $csv)])]
▶ 'a,a;b,b;foo,foo;bar,bar'
```

This code works, but it is a bit unreadable. In particular, since `str:split`
outputs multiple values but `each` wants a list argument, you have to wrap the
output of `str:split` in a list with `[(str:split ...)]`. Then you have to do
this again in order to pass the output of `each` to `str:join`. You might wonder
why commands like `str:split` and `each` do not simply output a list to make
this easier.

The answer to that particular question is in the next subsection, but for the
program at hand, there is a much better way to write it:

```elvish-transcript
~> var csv = a,b,foo,bar
~> use str
~> str:split , $csv | each {|x| put $x,$x } | str:join ';'
▶ 'a,a;b,b;foo,foo;bar,bar'
```

Besides having fewer pairs of parentheses (and brackets), this program is also
more readable, because the data flows from left to right, and there is no
nesting. You can see that `$csv` is first split by commas, then each value gets
reduplicated, and then finally everything is joined by semicolons. It matches
exactly how you would describe the algorithm in spoken English -- or for that
matter, any spoken language!

Both versions work, because commands like `each` and `str:join` that work with
multiple inputs can take their inputs in two ways: they can take the inputs as
one list argument, like in the first version; or from the pipeline, like the
second version. Whenever possible, you should prefer the input-from-pipeline
form: it makes for programs that have little nesting, read naturally.

One exception to the recommendation is when the input is a small set of things
known beforehand. For example:

```elvish-transcript
~> each $str:to-upper~ [lorem ipsum]
▶ LOREM
▶ IPSUM
```

Here, using the input-from-argument is completely fine: if you want to use the
input-from-input form, you have to supply the input using `put`, which is also
OK but a bit more wordy:

```elvish-transcript
~> put lorem ipsum | each $str:to-upper~
▶ LOREM
▶ IPSUM
```

However, not all commands support taking input from the pipeline. For example,
if we want to first join some values with space and then split at commas, this
won't work:

```elvish-transcript
~> use str
~> str:join ' ' [a,b c,d] | str:split ,
Exception: arity mismatch: arguments here must be 2 values, but is 1 value
[tty], line 1: str:join ' ' [a,b c,d] | str:split ,
```

This is because the `str:split` command only ever works with one input (one
string to split), and was not implemented to support taking input from pipeline;
hence it always takes 2 arguments and we got an exception.

It is easy to remedy this situation however. The `all` command passes its input
to its output, and by capturing its output, we can turn the input into an
argument:

```elvish-transcript
~> use str
~> str:join ' ' [a,b c,d] | str:split , (all)
▶ a
▶ 'b c'
▶ d
```

## Streaming Multiple Outputs

In the previous subsection, we remarked that commands like `str:split` and
`each` write multiple output values instead of one list. Why?

This has to do with another advantage of passing data through the pipeline: in a
pipeline, all commands are executed in parallel. A command in a pipeline does
not need to wait for its previous command to finish running before it can start
processing data. Try this in your terminal:

```elvish-transcript
~> each $str:to-upper~ | each {|x| put $x$x }
(Start typing)
abc
▶ ABCABC
xyz
▶ XYZXYZ
(Press ^D)
```

You will notice that as soon as you press Enter after typing `abc`, the output
`ABCABC` is shown. As soon as one input is available, it goes through the entire
pipeline, each command doing its work. This gives you immediate feedback, and
makes good use of multi-core CPUs on modern computers. Pipelines are like
assembly lines in the manufacturing industry.

If instead of passing multiple values, we pass a list through the pipeline: that
means that each command will now be waiting for its previous command to do all
the processing and pack the results in a list before it can start doing
anything. Now, although the commands themselves are run in parallel, they all
need to be waiting for their previous commands to finish before they can start
doing real work.

This is why commands like `each` and `str:split` produce multiple values instead
of one list. When writing your functions, try to make them produce multiple
values as well: they will cooperate better with builtin commands, and they can
benefit from the efficiency of parallel computations.

# Working with Multiple Values

In Elvish, many constructs can evaluate to multiple values. This can be
surprising if you are not familiar with it.

To start with, output captures evaluate to all the captured values, instead of a
list:

```elvish-transcript
~> use str
~> str:split , a,b,c
▶ a
▶ b
▶ c
~> var li = (str:split , a,b,c)
Exception: arity mismatch: assignment right-hand-side must be 1 value, but is 3 values
[tty], line 1: li = (str:split , a,b,c)
```

The assignment fails with "arity mismatch" because the right hand side evaluates
to 3 values, but you are attempting to assign them to just one variable. If you
want to capture the results into a list, you have to explicitly do so, either by
constructing a list or using rest variables:

```elvish-transcript
~> use str
~> var li = [(str:split , a,b,c)]
~> put $li
▶ [a b c]
~> var @li = (str:split , a,b,c) # equivalent and slightly shorter
```

## Assigning Multiple Variables

# To Be Continued...

As of writing, Elvish is neither stable nor complete. The builtin libraries
still have missing pieces, the package manager is in its early days, and things
like a type system and macros have been proposed and considered, but not yet
worked on. Deciding best practices for using feature *x* can be a bit tricky
when that feature *x* doesn't yet exist!

The current version of the document is what the lead developer of Elvish (@xiaq)
has collected as best practices for writing Elvish code in early 2018, between
the release of Elvish 0.11 and 0.12. They apply to aspects of the Elvish
language that are relatively complete and stable; but as Elvish evolves, the
document will co-evolve. You are invited to revisit this document once in a
while!
