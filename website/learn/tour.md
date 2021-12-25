<!-- toc number-sections -->

# Introduction

Welcome to the quick tour of Elvish. This tour works best if you have used
another shell or programming language before.

If you are mainly interested in using Elvish interactively, jump directly to
[interactive features](#interactive-features).

# Basic shell language

Many basic language features of Elvish are very familiar to traditional shells.
A notable exception is control structures, covered [below](#control-structures)
in the _advanced language features_ section.

## Comparison with bash

If you are familiar with bash, the following table shows some (rough)
correspondence between Elvish and bash syntax:

<table>
  <tr>
    <th>Feature</th>
    <th>Elvish</th>
    <th>bash equivalent</th>
  </tr>

  <tr>
    <td><a href="#barewords">Barewords</a></td>
    <td colspan="2"><code>echo foo</code></td>
  </tr>
  <tr>
    <td rowspan="2"><a href="#single-quoted-strings">Single-quoted strings</a></td>
    <td colspan="2"><code>echo 'foo'</code></td>
  </tr>
  <tr>
    <td><code>echo 'It''s good'</code></td>
    <td><code>echo 'It'\''s good'</code></td>
  </tr>
  <tr>
    <td rowspan="3"><a href="#double-quoted-strings">Double-quoted strings</a></td>
    <td colspan="2"><code>echo "foo"</code></td>
  </tr>
  <tr>
    <td><code>echo "foo\nbar"</code></td>
    <td><code>echo $'foo\nbar'</code></td>
  </tr>
  <tr>
    <td><code>echo "foo: "$foo</code></td>
    <td><code>echo "foo: $foo"</code></td>
  </tr>
  <tr>
    <td><a href="#comments">Comments</a></td>
    <td colspan="2"><code># comment</code></td>
  </tr>
  <tr>
    <td><a href="#line-continuation">Line continuation</a></td>
    <td><code>echo foo ^<br> bar</code></td>
    <td><code>echo foo \<br> bar</code></td>
  </tr>
  <tr>
    <td><a href="#brace-expansion">Brace expansion</a></td>
    <td><code>echo {foo bar}.txt</code></td>
    <td><code>echo {foo,bar}.txt</code></td>
  </tr>
  <tr>
    <td rowspan="3"><a href="#wildcards">Wildcards</a></td>
    <td colspan="2"><code>echo *.?</code></td>
  </tr>
  <tr>
    <td><code>echo **.go</code></td>
    <td><code>find . -name '*.go'</code></td>
  </tr>
  <tr>
    <td><code>echo *.?[set:ch]</code></td>
    <td><code>echo *.[ch]</code></td>
  </tr>
  <tr>
    <td><a href="#tilde-expansion">Tilde expansion</a></td>
    <td colspan="2"><code>echo ~/foo</code></td>
  </tr>
  <tr>
    <td rowspan="3"><a href="#setting-variables">Setting variables</a></td>
    <td><code>var foo = bar</code></td>
    <td><code>foo=bar</code></td>
  </tr>
  <tr>
    <td><code>set foo = bar</code></td>
    <td><code>foo=bar</code></td>
  </tr>
  <tr>
    <td colspan="2"><code>foo=bar cmd</code></td>
  </tr>
  <tr>
    <td rowspan="2"><a href="#using-variables">Using variables</a></td>
    <td colspan="2"><code>echo $foo</code></td>
  </tr>
  <tr>
    <td><code>echo $E:HOME</code></td>
    <td><code>echo $HOME</code></td>
  </tr>
  <tr>
    <td><a href="#redirections">Redirections</a></td>
    <td colspan="2"><code>head -n10 < a.txt > b.txt</code></td>
  </tr>
  <tr>
    <td><a href="#byte-pipelines">Byte pipelines</a></td>
    <td colspan="2"><code>head -n4 a.txt | grep x</code></td>
  </tr>
  <tr>
    <td><a href="#output-capture">Output capture</a></td>
    <td><code>ls -l (which elvish)</code></td>
    <td><code>ls -l $(which elvish)</code></td>
  </tr>
  <tr>
    <td><a href="#background-jobs">Background jobs</a></td>
    <td colspan="2"><code>echo foo &amp;</code></td>
  </tr>

</table>

## Barewords

Like traditional shells, unquoted words that don't contain special characters
are treated as strings (such words are called **barewords**):

```elvish-transcript
~> echo foobar
foobar
~> ls /
bin   dev  home  lib64       mnt  proc  run   srv  tmp  var
boot  etc  lib   lost+found  opt  root  sbin  sys  usr
~> vim a.c
```

This is one of the most distinctive syntactical features of shells; non-shell
programming languages typically treat unquoted words as names of functions and
variables.

Read the language reference on [barewords](../ref/language.html#bareword) to
learn more.

## Single-quoted strings

Like traditional shells, single-quoted strings expand nothing; every character
represents itself (except the single quote itself):

```elvish-transcript
~> echo 'hello\world$'
hello\world$
```

Like [Plan 9 rc](http://doc.cat-v.org/plan_9/4th_edition/papers/rc), or zsh with
the `RC_QUOTES` option turned on, the single quote itself can be written by
doubling it

```elvish-transcript
~> echo 'it''s good'
it's good
```

Read the language reference on
[single-quoted strings](../ref/language.html#single-quoted-string) to learn
more.

## Double-quoted strings

Like many non-shell programming languages and `$''` in bash, double-quoted
strings support C-like escape sequences (e.g. `\\n` for newline):

```elvish-transcript
~> echo "foo\nbar"
foo
bar
```

Unlike traditional shells, Elvish does **not** support interpolation inside
double-quoted strings. Instead, you can just write multiple words together, and
they will be concatenated:

```elvish-transcript
~> var x = foo
~> put 'x is '$x
▶ 'x is foo'
```

Read the language reference on
[double-quoted strings](../ref/language.html#double-quoted-string) to learn
more.

## Comments

Comments start with `#` and extend to the end of the line:

```elvish-transcript
~> echo foo # this is a comment
foo
```

## Line continuation

Line continuation in Elvish uses `^` instead of `\`:

```elvish-transcript
~> echo foo ^
   bar
foo bar
```

Unlike traditional shells, line continuation is treated as whitespace. In
Elvish, the following code outputs `foo bar`:

```elvish
echo foo^
bar
```

However, in bash, the following code outputs `foobar`:

```bash
echo foo\
bar
```

## Brace expansion

Brace expansions in Elvish work like in traditional shells, but use spaces
instead of commas:

```elvish-transcript
~> echo {foo bar}.txt
foo.txt bar.txt
```

The opening brace `{` **must not** be followed by a whitespace, to disambiguate
from [lambdas](#lambdas).

**Note**: commas might still work as a separator in Elvish's brace expansions,
but it will eventually be deprecated and removed soon.

Read the language reference on [braced lists](../ref/language.html#braced-list)
to learn more.

## Wildcards

The basic wildcard characters, `*` and `?`, work like in traditional shells:

```elvish-transcript
~> ls
bar.ch  d1  d2  d3  foo.c  foo.h  lorem.go  lorem.txt
~> echo *.?
foo.c foo.h
```

Elvish also supports `**`, which matches multiple path components:

```elvish-transcript
~> find . -name '*.go'
./d1/a.go
./d2/b.go
./lorem.go
./d3/d4/c.go
~> echo **.go
d1/a.go d2/b.go d3/d4/c.go lorem.go
```

Character classes are a bit more verbose in Elvish. First, a character set is
written like `[set:ch]`, instead of just `[ch]`. Second, they don't appear on
their own, but as a suffix to `?`. For example, to match files ending in either
`.c` or `.h`, use:

```elvish-transcript
~> echo *.?[set:ch]
foo.c foo.h
```

A character set suffix can also be applied to `*`. For example, to match files
who extension only contains `c` and `h`:

```elvish-transcript
~> echo *.*[set:ch]
bar.ch foo.c foo.h
```

Read the language reference on
[wildcard expansion](../ref/language.html#wildcard-expansion) to learn more.

## Tilde expansion

Tilde expansion works likes in traditional shells. Assuming that the home
directory of the current user is `/home/me`, and the home directory of `elf` is
`/home/elf`:

```elvish-transcript
~> echo ~/foo
/home/me/foo
~> echo ~elf/foo
/home/elf/foo
```

Read the language reference on
[tilde expansion](../ref/language.html#tilde-expansion) to learn more.

## Setting variables

Variables are declared with the `var` command, and set with the `set` command:

```elvish-transcript
~> var foo = bar
~> echo $foo
bar
~> set foo = quux
~> echo $foo
quux
```

The spaces around `=` are mandatory.

Unlike traditional shells, variables must be declared before they can be set;
setting an undeclared variable results in an error.

Like traditional shells, Elvish supports setting a variable temporarily for the
duration of a command, by prefixing the command with `foo=bar`. For example:

```elvish-transcript
~> var foo = original
~> fn f { echo $foo }
~> foo=temporary f
temporary
~> echo $foo
original
```

Read the language reference on [the `var` command](../ref/language.html#var),
[the `set` command](../ref/language.html#set) and
[temporary assignments](../ref/language.html#temporary-assignment) to learn
more.

## Using variables

Like traditional shells, using the value of a variable requires the `$` prefix.

```elvish-transcript
~> var foo = bar
~> echo $foo
bar
```

Unlike traditional shells, variables must be declared before being used; if the
`foo` variable wasn't declared with `var` first, `echo $foo` results in an
error.

Also unlike traditional shells, environment variables in Elvish live in a
separate `E:` namespace:

```elvish-transcript
~> echo $E:HOME
/home/elf
```

Read the language reference on [variables](../ref/language.html#variable) and
[special namespaces](../ref/language.html#special-namespaces) to learn more.

## Redirections

Redirections in Elvish work like in traditional shells. For example, to save the
first 10 lines of `a.txt` to `a1.txt`:

```elvish-transcript
~> head -n10 < a.txt > a1.txt
```

Read the language reference on [redirections](../ref/language.html#redirection)
to learn more.

## Byte pipelines

UNIX pipelines in Elvish (called **byte pipelines**, to distinguish from
[value pipelines](#value-pipelines)) work like in traditional shells. For
example, to find occurrences of `x` in the first 4 lines of `a.txt`:

```elvish-transcript
~> cat a.txt
foo
barx
lorem
quux
lux
nox
~> head -n4 a.txt | grep x
barx
quux
```

Read the language reference on [pipelines](../ref/language.html#pipeline) to
learn more.

## Output capture

Output of commands can be captured and used as values with `()`. For example,
the following command shows details of the `elvish` binary:

```elvish-transcript
~> ls -l (which elvish)
-rwxr-xr-x 1 xiaq users 7813495 Mar  2 21:32 /home/xiaq/go/bin/elvish
```

**Note**: the same feature is usually known as _command substitution_ in
traditonal shells.

Unlike traditional shells, Elvish only splits the output on newlines, not any
other whitespace characters.

Read the language reference on
[output capture](../ref/language.html#output-capture) to learn more.

## Background jobs

Add `&` to the end of a pipeline to make it run in the background, similar to
traditional shells:

```elvish-transcript
~> echo foo &
foo
job echo foo & finished
```

Unlike traditional shells, the `&` character does not serve to separate
commands. In bash you can write `echo foo & echo bar`; in Elvish you still need
to terminate the first command with `;` or newline: `echo foo &; echo bar`.

Read the language reference on
[background pipelines](../ref/language.html#background-pipeline) to learn more.

# Advanced language features

Building on a core of familiar shell-like syntax, the Elvish language
incorporates many advanced features that make it a modern dynamic programming
language.

## Value output

Like in traditional shells, commands in Elvish can output **bytes**. The `echo`
command outputs bytes:

```elvish-transcript
~> echo foo bar
foo bar
```

Additionally, commands can also output **values**. Values include not just
strings, but also [lambdas](#lambdas), [numbers](#numbers),
[lists and maps](#lists-and-maps). The `put` command outputs values:

```elvish-transcript
~> put foo [foo] [&foo=bar] { put foo }
▶ foo
▶ [foo]
▶ [&foo=bar]
▶ <closure 0xc000347500>
```

Many builtin commands output values. For example, string functions in the `str:`
module outputs their results as values. This makes those functions work
seamlessly with strings that contain newlines or even NUL bytes:

```elvish-transcript
~> use str
~> str:join ',' ["foo\nbar" "lorem\x00ipsum"]
▶ "foo\nbar,lorem\x00ipsum"
```

Unlike most programming languages, Elvish commands don't have return values.
Instead, they use the value output to "return" their results.

Read the reference for [builtin commands](../ref/builtin.html) to learn which
commands work with value inputs and outputs. Among them, here are some
general-purpose primitives:

<table>
  <tr>
    <th>Command</th>
    <th>Functionality</th>
  </tr>
  <tr>
    <td>
      <a href="../ref/builtin.html#all"><code>all</code></a>
    </td>
    <td>
      Passes value inputs to value outputs
    </td>
  </tr>
  <tr>
    <td>
      <a href="../ref/builtin.html#each"><code>each</code></a>
    </td>
    <td>
      Applies a function to all values from value input
    </td>
  </tr>
  <tr>
    <td>
      <a href="../ref/builtin.html#put"><code>put</code></a>
    </td>
    <td>
      Writes arguments as value outputs
    </td>
  </tr>
  <tr>
    <td>
      <a href="../ref/builtin.html#slurp"><code>slurp</code></a>
    </td>
    <td>
      Convert byte input to a single string in value output
    </td>
  </tr>
</table>

## Value pipelines

Pipelines work with value outputs too. When forming pipelines, a command that
writes value outputs can be followed by a command that takes value inputs. For
example, the `each` command takes value inputs, and applies a lambda to each one
of them:

```elvish-transcript
~> put foo bar | each {|x| echo 'I got '$x }
I got foo
I got bar
```

Read the language reference on [pipelines](../ref/language.html#pipeline) to
learn more about pipelines in general.

## Value output capture

Output capture works with value output too. Capturing value outputs always
recovers the exact values there were written. For example, the `str:join`
command joins a list of strings with a separator, and its output can be captured
and saved in a variable:

```elvish-transcript
~> use str
~> var s = (str:join ',' ["foo\nbar" "lorem\x00ipsum"])
~> put $s
▶ "foo\nbar,lorem\x00ipsum"
```

Read the language reference on
[output capture](../ref/language.html#output-capture) to learn more.

## Lists and maps

Lists look like `[a b c]`, and maps look like `[&key1=value1 &key2=value2]`:

```elvish-transcript
~> var li = [foo bar lorem ipsum]
~> put $li
▶ [foo bar lorem ipsum]
~> var map = [&k1=v2 &k2=v2]
~> put $map
▶ [&k1=v2 &k2=v2]
```

You can get elements of lists and maps by indexing them. Lists are zero-based
and support slicing too:

```elvish-transcript
~> put $li[0]
▶ foo
~> put $li[1..3]
▶ [bar lorem]
~> put $map[k1]
▶ v2
```

Read the language reference on [lists](../ref/language.html#list) and
[maps](../ref/language.html#map) to learn more.

## Numbers

Elvish has a double-precision floating-point number type, `float64`. There is no
dedicated syntax for it; instead, it can constructed using the `float64`
builtin:

```elvish-transcript
~> float64 1
▶ (float64 1)
~> float64 1e2
▶ (float64 100)
```

Most arithmetic commands in Elvish support both typed numbers and strings that
can be converted to numbers. They usually output typed numbers:

```elvish-transcript
~> + 1 2
▶ (float64 3)
~> use math
~> math:pow (float64 10) 3
▶ (float64 1000)
```

**Note**: The set of number types will likely expand in future.

Read the language reference on [numbers](../ref/language.html#number) and the
reference for the [math module](../ref/math.html) to learn more.

## Booleans

Elvish has two boolean values, `$true` and `$false`.

Read the language reference on [booleans](../ref/language.html#boolean) to learn
more.

## Options

Many Elvish commands take **options**, which look like map pairs (`&key=value`).
For example, the `echo` command takes a `sep` option that can be used to
override the default separator of space:

```elvish-transcript
~> echo &sep=',' foo bar
foo,bar
~> echo &sep="\n" foo bar
foo
bar
```

## Lambdas

Lambdas are first-class values in Elvish. They can be saved in variables, used
as commands, passed to commands, and so on.

Lambdas can be written by enclosing its body with `{` and `}`:

```elvish-transcript
~> var f = { echo "I'm a lambda" }
~> $f
I'm a lambda
~> put $f
▶ <closure 0xc000265bc0>
~> var g = (put $f)
~> $g
I'm a lambda
```

The opening brace `{` **must** be followed by some whitespace, to disambiguate
from [brace expansion](#brace-expansion).

Lambdas can take arguments and options, which can be written in a **signature**:

```elvish-transcript
~> var f = {|a b &opt=default|
     echo "a = "$a
     echo "b = "$b
     echo "opt = "$opt
   }
~> $f foo bar
a = foo
b = bar
opt = default
~> $f foo bar &opt=option
a = foo
b = bar
opt = option
```

There must be no space between the `]` and `{` in this case.

Read the language reference on [functions](../ref/language.html#function) to
learn more about functions.

## Control structures

Control structures in Elvish look very different from traditional shells. For
example, this is how an `if` command looks:

```elvish-transcript
~> if (eq (uname) Linux) { echo "You're on Linux" }
You're on Linux
```

The `if` command takes a conditional expression (an output capture in this
case), and the body to execute as a lambda. Since lambdas allow internal
newlines, you can also write it like this:

```elvish-transcript
~> if (eq (uname) Linux) {
     echo "You're on Linux"
   }
You're on Linux
```

However, you must write the opening brace `{` on the same line as `if`. If you
write it on a separate line, Elvish would parse it as two separate commands.

The `for` command looks like this:

```elvish-transcript
~> for x [expressive versatile] {
     echo "Elvish is "$x
   }
Elvish is expressive
Elvish is versatile
```

Read the language reference on [the `if` command](../ref/language.html#if),
[the `for` command](../ref/language.html#for), and additionally
[the `while` command](../ref/language.html#while) to learn more.

## Exceptions

Elvish uses exceptions to signal errors. For example, calling a function with
the wrong number of arguments throws an exception:

```elvish-transcript
~> var f = { echo foo } # doesn't take arguments
~> $f a b
Exception: arity mismatch: arguments here must be 0 values, but is 2 values
[tty 2], line 1: $f a b
```

Moreover, non-zero exits from external commands are also turned into exceptions:

```elvish-transcript
~> false
Exception: false exited with 1
[tty 3], line 1: false
```

Exceptions can be caught using the `try` command:

```elvish-transcript
~> try {
     false
   } except e {
     echo 'got an exception'
   }
got an exception
```

Read the language reference on
[the exception value type](../ref/language.html#exception) and
[the `try` command](../ref/language.html#try) to learn more.

## Namespaces and modules

The names of variables and functions can have **namespaces** prepended to their
names. Namespaces always end with `:`.

The [using variables](#using-variables) section has already shown the `E:`
namespace. Other namespaces can be added by importing modules with `use`. For
example, [the `str:` module](../ref/str.html) provides string utilities:

```elvish-transcript
~> use str
~> str:to-upper foo
▶ FOO
```

You can define your own modules by putting `.elv` files in
`~/.config/elvish/lib` (or `~\AppData\Roaming\elvish\lib`). For example, to
define a module called `foo`, put the following in `foo.elv` under the
aforementioned directory:

```elvish
fn f {
  echo 'in a function in foo'
}
```

This module can now be used like this:

```elvish-transcript
~> use foo
~> foo:f
in a function in foo
```

Read the language reference on
[namespaces and modules](../ref/language.html#namespaces-and-modules) to learn
more.

## External command support

As shown in examples above, Elvish supports calling external commands directly
by writing their name. If an external command exits with a non-zero code, it
throws an exception.

Unfortunately, many of the advanced language features are only available for
internal commands and functions. For example:

-   They can only write byte output, not value output.

-   They only take string arguments; non-string arguments are implicitly coerced
    to strings.

-   They don't take options.

Read the language reference on
[ordinary commands](../ref/language.html#ordinary-command) to learn more about
when Elvish decides that a command is an external command.

# Interactive features

Read [the API of the interactive editor](../ref/edit.html) to learn more about
UI customization options.

## Tab completion

Press <span class="key">Tab</span> to start completion. For example, after
typing `vim` and <span class="key">Space</span>, press
<span class="key">Tab</span> to complete filenames:

@ttyshot tour/completion

Basic operations should be quite intuitive:

-   To navigate the candidate list, use arrow keys <span class="key">▲</span>
    <span class="key">▼</span> <span class="key">◀</span>
    <span class="key">▶</span> or <span class="key">Tab</span> and
    <span class="key">Shift-Tab</span>.

-   To accept the selected candidate, press <span class="key">Enter</span>.

-   To cancel, press <span class="key">Escape.</span>

As indicated by the horizontal scrollbar, you can scroll to the right to find
additional results that don't fit in the terminal.

You may have noticed that the cursor has moved to the right of "COMPLETING
argument". This indicates that you can continue typing to filter candidates. For
example, after typing `.md`, the UI looks like this:

@ttyshot tour/completion-filter

Read the reference on [completion API](../ref/edit.html#completion-api) to learn
how to program and customize tab completion.

## Command history

Elvish has several UI features for working with command history.

### History walking

Press <span class="key">▲</span> to fetch the last command. This is called
**history walking** mode:

@ttyshot tour/history-walk

Press <span class="key">▲</span> to go further back, <span class="key">▼</span>
to go forward, or <span class="key">Escape</span> to cancel.

To restrict to commands that start with a prefix, simply type the prefix before
pressing <span class="key">▲</span>. For example, to walk through commands
starting with `echo`, type `echo` before pressing <span class="key">▲</span>:

@ttyshot tour/history-walk-prefix

### History listing

Press <span class="key">Ctrl-R</span> to list the full command history:

@ttyshot tour/history-list

Like in completion mode, type to filter the list, press
<span class="key">▲</span> and <span class="key">▼</span> to navigate the list,
<span class="key">Enter</span> to insert the selected entry, or <span
class="key">Escape</span> to cancel.

### Last command

Finally, Elvish has a **last command** mode dedicated to inserting parts of the
last command. Press <span class="key">Alt-,</span> to trigger it:

@ttyshot tour/lastcmd

## Directory history

Elvish remembers which directories you have visited. Press <span
class="key">Ctrl-L</span> to list visited directories. Use
<span class="key">▲</span> and <span class="key">▼</span> to navigate the list,
<span class="key">Enter</span> to change to that directory, or
<span class="key">Escape</span> to cancel.

@ttyshot tour/location

Type to filter:

@ttyshot tour/location-filter

## Navigation mode

Press <span class="key">Ctrl-N</span> to start the builtin filesystem navigator,
or **navigation mode**.

@ttyshot tour/navigation

Unlike other modes, the cursor stays in the main buffer in navigation mode. This
allows you to continue typing commands; while doing that, you can press
<span class="key">Enter</span> to insert the selected filename. You can also
press <span class="key">Alt-Enter</span> to insert the filename without exiting
navigation mode; this is useful when you want to insert multiple filenames.

## Startup script

Elvish's interactive startup script is [`rc.elv`](../ref/command.html#rc-file).
Non-interactive Elvish sessions do not have a startup script.

### POSIX aliases

Elvish doesn't support POSIX aliases, but you can get a similar experience
simply by defining functions:

```elvish
fn ls {|@a| e:ls --color $@a }
```

The `e:` prefix (for "external") ensures that the external command named `ls`
will be called. Otherwise this definition will result in infinite recursion.

### Prompt customization

The left and right prompts can be customized by assigning functions to
`edit:prompt` and `edit:rprompt`. The following configuration simulates the
default prompts, but uses fancy Unicode:

```elvish
# "tilde-abbr" abbreviates home directory to a tilde.
set edit:prompt = { tilde-abbr $pwd; put '❱ ' }
# "constantly" returns a function that always writes the same value(s) to
# output; "styled" writes styled output.
set edit:rprompt = (constantly (styled (whoami)✸(hostname) inverse))
```

This is how it looks:

@ttyshot tour/unicode-prompts

### Changing PATH

Another common task in the interactive startup script is to set the search path.
You can do set the environment variable directly (all environment variables have
a `E:` prefix):

```elvish
set E:PATH = /opts/bin:/bin:/usr/bin
```

But it is usually nicer to set the [`$paths`](../ref/builtin.html#paths)
instead:

```elvish
set paths = [/opts/bin /bin /usr/bin]
```
