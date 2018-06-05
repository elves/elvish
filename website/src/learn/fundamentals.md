<!-- toc -->

**This tutorial is quite incomplete, and it is being constantly expanded.**

This tutorial introduces the fundamentals of shell programming with Elvish.
It does not assume familiarity with other shells, but some understanding of
basic programming concepts is required.

This tutorial evolves around a "hello" program. But if your primary interest
in shell programming is not saying hello (which, by the way, is a shame), the
tutorial also contains some hints and examples on how to apply to knowledge to
non-hello applications.

**Note**: Elvish is very similar to other shell languages in many aspects, but
also very different in others. When transferring your knowledge of Elvish to
another shell, it is worthwhile to first check how things work there.


# Hello, world!

Let's begin with the most traditional "hello world" program. In Elvish,
you invoke the `echo` **command** to print something on the terminal:

```elvish-transcript
~> echo "Hello, world!"
Hello, world!
```

In Elvish, as in other shells, command invocations follow a simple structure: you
write the command name, followed by arguments, all separated by spaces (or
tabs). No parentheses or commas are needed.

We enclose our text here in double quotes, making it a **string literal**.
Compared to other languages, shell languages are a bit sloppy in that they
allow you to write strings *without* quotes. The following also works:

```elvish-transcript
~> echo Hello, world!
Hello, world!
```

However, the way it works has a subtle difference: here `Hello,` and `world!`
are two arguments (remember that spaces separate arguments), and `echo` joins
them together with a space. This is apparent if you put multiple spaces
between them:

```elvish-transcript
~> echo Hello,      world!
Hello, world!
~> echo "Hello,     world!"
Hello,     world!
```

When you write your message without quotes, no matter how many spaces there
are, it is always the same two arguments `Hello,` and `world!`. If you quote
your message, the spaces are part of the string and thus preserved.

It is a good idea to always quote your string when it contains spaces or any
special symbols other than period (`.`), dash (`-`) or underscore (`_`).


## It Doesn't Have to be Hello

All command invocations in shell have the same basic structure: name of
command, followed by arguments. Elvish provides a lot of useful [builtin
commands](/ref/builtin.html), and `echo` is just one of them. As another
example, there is one for generating random numbers, which you can use as a
digital dice:

```elvish-transcript
~> # randint a b generates an integer from range a...b-1
   randint 1 7
▶ 3
```

Arithmetic operations are also commands. Since they also follow the same order
of command name first, the syntax deviates a bit from usual mathematical
notations:

```elvish-transcript
~> * 17 28 # multiplication
▶ 476
~> ^ 2 10 # exponention
▶ 1024
```

The commands introduced above -- `echo`, `randint`, `*` and `^` -- are all
**builtin** commands, commands that Elvish provides for you.

As a shell language, however, Elvish also makes it trivial to use **external**
commands, commands implemented as separate programs. Chances are you have
already used some of them like `ls` or `cat`. Here we show you how to obtain
Elvish entirely from the command line: you can use `wget` to download files,
`shasum` to verify its checksum, and `tar` to uncompress them, all of which
are external commands:

```elvish-transcript
~> wget https://dl.elvish.io/elvish-linux.tar.gz
... omit ...
elvish-linux.tar.gz  100%[======================>]   4.91M  10.9MB/s    in 0.4s
~> shasum -a 256 elvish-linux.tar.gz
0fc3c145a81345a1c49576b86ef12156d4eba1829e1bb20e9c39d115991a9c7b elvish-linux.tar.gz
~> tar xvf elvish-linux.tar.gz
x elvish
```

With the most basic knowledge of how to invoke commands, there is already a
myriad of functionalities at your fingertip.


# Hello, {insert user name}!

The "hello world" program is a classic, but the fact that it always prints the
same simple message does make it a little bit boring.

One way to make programs more useful is to teach them to do different things
depending on the context. In the case of our "hello world" program, why not
teach it to greet whoever is running the program?

```elvish-transcript
~> echo "Hello, world! Hello, "$E:USER"!"
Hello, world! Hello, xiaq!
```

There are several things happening here. First, `$E:USER` represents the `USER`
**environment variable** ("E" being mnemonic for "environment"). In UNIX
environments, it is usually set to the name of the current user. Second, we
are running several strings and a variable all together: in this case, Elvish
will concatenate them for you. Hence the result we see.

Depending on your taste, you might feel that it's nicer to greet the world and
the user on separate lines. To do this, we can insert `\n`, an **escape
sequence** representing a newline in our string:

```elvish-transcript
~> echo "Hello, world!\nHello, "$E:USER"!"
Hello, world!
Hello, xiaq!
```

There are many such sequences starting with a backslash, including `\\` which
represents the backslash itself. Beware that such escape sequences only work
within double quotes.

Environment variables are not the only way to learn about a computer system;
we can also gain more information by invoking commands. For instance, the
`uname` command tells you which operation system the computer is running:

```elvish-transcript
~> uname
Darwin
```

([Darwin](https://en.wikipedia.org/wiki/Darwin_(operating_system)), by the
way, is the open-source core of macOS and iOS.)

To incorporate the output of `uname` into our hello message, we can first
**capture** its output using parentheses and keep it in a variable using the
assignment form `variable = value`:

```elvish-transcript
~> os = (uname)
~> # Variable "os" now contains "Darwin"
```

We can then use this variable, in a similar way to how we used the environment
variable, just without the `E:` namespace:

```elvish-transcript
~> echo "Hello, "$os" user!"
Hello, Darwin user!
```

I used a variable for demonstration, but it's also possible to forego the
variable and use the captured output directly:

```elvish-transcript
~> echo "Hello, "(uname)" user!"
Hello, Darwin user!
```

## It Doesn't Have to be Hello

Output captures can get you quite far in combining commands. For instance, you
can use output captures to construct do complex arithmetic involving more than
one operation:

```elvish-transcript
~> # compute the answer to life, universe and everything
   * (+ 3 4) (- 100 94)
▶ 42
```


<!--
# Hello, everyone!


Now let's say you want to say hello to several people, and typing `Hello` repeatedly is tiresome. You can save some work by using a **for-loop**:

```elvish
for name [Julius Pompey Marcus] {
    echo 'Hello, '$name'!'
}
```

In elvish you can put newlines between the elements to loop over, as long as they are terminated by `; do`.

For easier reuse, you can also create a **list** to store the names:

```elvish
triumvirate = [Julius Pompey Marcus]
```

Lists are surrounded by square brackets, like in several other languages. Elements are separated by whitespaces.

As you may have noticed, dashes are allowed in variable names. You are encouraged to use them instead of underscores; they are easier to type and more readable (after a little getting-used-to).

Now it's time to use our list of the first triumvirate:

```elvish
for name in $first-triumvirate; do
    echo 'Hello, '$name'!'
done
```

This will, however, results in an error, saying that a string and a list cannot be concatenated. Why? Remember that `$x` is always one value. This is even true for lists, so the `for` loop only sees one value to loop over, namely the list itself.

To make multiple words out of a list, you must explicitly **splice** the list with an `@` before the variable name:

```elvish
for name in $@first-triumvirate; do
    echo 'Hello, '$name'!'
done
```

# Each person gets $hello~'ed

The for-loop we just show can also be written in a functional style:

```elvish
each [name]{
    echo 'Hello, '$name'!'
} $first-triumvirate
```

This looks similar to the for-loop version, but it makes use of a remarkable construct -- an **anonymous function**, also known as a **lambda**. In elvish, a lambda is syntactically formed by an argument list followed immediately (without space) by a function body enclosed in braces. Here, `[name]{ echo 'Hello, '$name'!' }` is a lambda that takes exactly one argument and calls `echo` to do the helloing. We pass it along a list to the `each` builtin, which runs the function on each element of the list.

Functions, like strings and lists, can be stored in variables:

```elvish
hello=[name]{ echo 'Hello, '$name'!' }
each $hello $first-triumvirate
```

To call a function, simply use it as a command:

```elvish
$hello 'Mark Antony' # Hello, Mark Anthony!
```

You must have noticed that you have to use `$hello` instead of `hello` to call the function. This is because the *hello-the-variable* and *hello-the-command* are different enitites. To define new commands, use the `fn` special form:

```elvish
fn hello [name]{
    echo 'Hello, '$name'!'
}
hello Cicero # Hello, Cicero!
```

Users of traditional shells and Common Lisp will find this separation of the variable namespace and command namespace familiar.

However, in elvish this separation is only superficial; what `fn hello` really does is just defining a variable called `hello~`. You can prove this:

```elvish
echo $hello~ # <closure ...>
$hello~ Brutus # Hello, Brutus!
each $hello~ $first-triumvirate # (Hello to the first triumvirate)
```

Conversely, defining a variable `hello~` will also create a command named `hello`:

```elvish
hello~ = [name]{ echo "Hello, hello, "$name"!" }
hello Augustus # Hello, Augustus!
```

<!--
```
What I want to get into this document:

[ ] Command substitution

[ ] Rich pipeline

[X] Lists

[ ] Maps

[X] Lambdas

[X] fn

[X] $&

[X] One variable, one argument

[X] String syntax

[X] Lack of interpolation

[X] Several builtins -- each println

[ ] Editor API

[ ] Exception and verdict

[X] E: namespace for environment variables

[ ] e: namespace for external commands

[ ] Modules

Write for readers with a moderate knowledge of a POSIXy shell (bash, zsh, ...)
-->

