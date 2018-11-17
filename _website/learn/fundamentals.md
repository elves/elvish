<!-- toc number-sections -->

**Work in progress.**

This tutorial introduces the fundamentals of shell programming with Elvish.
Familiarity with other shells or programming languages is useful but not
required.


# Commands and strings

Let's begin with the traditional "hello world" program:

```elvish-transcript
~> echo Hello, world!
Hello, world!
```

Here, we call the `echo` **command** with two **arguments**: `Hello,` and
`world!`. The `echo` command prints them both, inserting a space in between
and adding a newline at the end, and voilà, we get back the message `Hello,
world!`.

## Quoting the argument

We used a single space between the two arguments. Using more also works:

```elvish-transcript
~> echo Hello,  world!
Hello, world!
```

The output still only has one space, because `echo` didn't know how many spaces
were used to separate the two arguments; all it sees is the two arguments,
`Hello,` and `world!`. However, you can preserve the two spaces by **quoting**
the entire text:

```elvish-transcript
~> echo "Hello,  world!"
Hello,  world!
```

In this version, `echo` only sees one argument containing
<code>Hello,&nbsp;&nbsp;world</code> (with two spaces). A pair of **double
quotes** tells Elvish that the text inside it is a single argument; the quotes
themselves are not part of the argument.

On contrary, the `Hello,` and `world!` arguments are implicitly delimited by
spaces (instead of explicitly by quotes); as such, they are known as
**barewords**. Some special characters also delimit barewords, which we will
see later. Barewords are useful to write command names, filenames and
command-line switches, which usually do not contain spaces or special
characters.


## Editing the command line

TODO

## Builtin and external commands

We demonstrated the basic command structure using `echo`, a very simple
command. The same structure applies to all the commands. For instance, Elvish
comes with a `randint` command that takes two arguments `a` and `b` and
generates a random integer in the range a...b-1. You can use the command as a
digital dice:

```elvish-transcript
~> randint 1 7
▶ 3
```

Arithmetic operations are also commands. Like other commands, the command
names comes first, making the syntax a bit different from common mathematical
notations:

```elvish-transcript
~> * 17 28 # multiplication
▶ 476
~> ^ 2 10 # exponention
▶ 1024
```

The commands above all come with Elvish; they are **builtin commands**. There
are [many more](../ref/builtin.html) of them.

Another kind of commands is **external commands**. They are separate programs
from Elvish, and either come with the operating system or are installed by you
manually. Chances are you have already used some of them, like `ls` for
listing files, or `cat` for showing files.

There are really a myriad of external commands; to start with, you can manage
code repositories with [git](https://git-scm.com), convert documents with
[pandoc](http://pandoc.org), process images with
[ImageImagick](https://www.imagemagick.org/script/index.php), transcode videos
with [ffmpeg](http://ffmpeg.org), test the security of websites with
[nmap](https://nmap.org) and analyze network traffic with
[tcpdump](http://www.tcpdump.org). Many free and open-source software come
with a command-line interface.

Here we show you how to obtain the latest version of Elvish entirely from the
command line: we use `curl` to download the binary and its checksum, `shasum`
to check the checksum, and `chmod` to make it executable (assuming that you are
running macOS on x86-64):

```elvish-transcript
~> curl -s -o elvish https://dl.elv.sh/darwin-amd64/elvish-HEAD
~> curl -s https://dl.elv.sh/darwin-amd64/elvish-HEAD.sha256sum
8b3db8cf5a614d24bf3f2ecf907af6618c6f4e57b1752e5f0e2cf4ec02bface0  elvish-HEAD
~> shasum -a 256 elvish
8b3db8cf5a614d24bf3f2ecf907af6618c6f4e57b1752e5f0e2cf4ec02bface0  elvish
~> chmod +x elvish
~> ./elvish
```

## History and scripting

Some commands are useful to rerun; for instance, you may want to roll the
digital dice several times in a roll, or for another occasion. Of course, you
can just retype the command:

```elvish-transcript
~> randint 1 7
▶ 1
```

The command is short, but still, it can become a chore if you want to run it
repeatedly. Fortunately, Elvish remembers all the commands you have typed;
you can just ask Elvish to recall it by pressing <span class="key">Up</span>:

$ttyshot fundamentals/history-1

This will give you the last command you have run. However, it may have been a
while when you have last run the `randint` command, and this will not give you
what you need. You can either continue pressing <span class="key">Up</span>
until you find the command, or you you can give Elvish a hint by typing some
characters from the command line you want, e.g. `ra`, before pressing <span
class="key">Up</span>:

$ttyshot fundamentals/history-2

Another way to rerun commands is saving them in a **script**, which is simply
a text file containing the commands you want to run. Using your favorite text
editor, save the command to `dice.elv` under your home directory:

```elvish
# dice.elv
randint 1 7
```

After saving the script, you can run it with:

```elvish-transcript
~> elvish dice.elv
▶ 4
```

Since the above command runs `elvish` explicitly, it works in other shells as
well, not just from Elvish itself.


# Variables and lists

To change what a command does, we now need to change the commands themselves.
For instance, instead of saying "Hello, world!", we might want our command to
say "Hello, John!":

```elvish-transcript
~> echo Hello, John!
Hello, John!
```

Which works until you want a different message. One way to solve this is using
**variables**:

```elvish-transcript
~> name = John
~> echo Hello, $name!
Hello, John!
```

The command `echo Hello, $name!` uses the `$name` variable you just assigned
in the previous command. To greet a different person, you can just change the
value of the variable, and the command doesn't need to change:

```elvish-transcript
~> name = Jane
~> echo Hello, $name!
Hello, Jane!
```

Using variables has another advantage: after defining a variable, you can use
it as many times as you want:

```elvish-transcript
~> name = Jane
~> echo Hello, $name!
Hello, Jane!
~> echo Bye, $name!
Bye, Jane!
```

Now, if you change the value of `$name`, the output of both commands will
change.


## Environment variables

In the examples above, we have assigned value of `$name` ourselves. We can
also make the `$name` variable automatically take the name of the current
user, which is usually kept in an **environment variable** called `USER`. In
Elvish, environment variables are used like other variables, except that they
have an `E:` at the front of the name:

```elvish-transcript
~> echo Hello, $E:USER!
Hello, elf!
~> echo Bye, $E:USER!
Bye, elf!
```

The outputs will likely differ on your machine.

## Lists and indexing

The values we have stored in variables so far are all strings. It is possible
to store a **list** of values in one variable; a list can be written by
surrounding some values with `[` and `]`. For example:

```elvish-transcript
~> list = [linux bsd macos windows]
~> echo $list
[linux bsd macos windows]
```

Each element of this list has an **index**, starting from 0. In the list
above, the index of `linux` is 0, that of `bsd` is 1, and so on. We can
retrieve an element by writing its index after the list, also surrounded by
`[` and `]`:

```elvish-transcript
~> echo $list[0] is at index 0
linux is at index 0
```

We can even do:

```elvish-transcript
~> echo [linux bsd macos windows][0] is at index 0
linux is at index 0
```

Note that in this example, the two pairs of `[]` have different meanings: the
first pair denotes lists, while the second pair denotes an indexing operation.

## Script arguments

Recall the `dice.elv` script above:

```elvish
# dice.elv
randint 1 7
```

And how we ran it:

```elvish-transcript
~> elvish dice.elv
▶ 4
```

We were using `elvish` itself as a command, with the sole argument `dice.elv`.
We can also supply additional arguments:

```elvish-transcript
~> elvish dice.elv a b c
▶ 4
```

But this hasn't made any difference, because well, our `dice.elv` script
doesn't make use of the arguments.

The arguments are kept in a `$args` variable, as a list. Let's try put this
into a `echo-args.elv` file in your home directory:

```elvish
echo $args
```

And we can run it:

```elvish-transcript
~> elvish show-args.elv
[]
~> elvish show-args.elv foo
[foo]
~> elvish show-args.elv foo bar
[foo bar]
```

Since `$args` is a list, we can retrieve the individual elements with
`$args[0]`, `$args[1]`, etc.. Let's rewrite our greet-and-bye script, taking
the name as an argument. Put this in `greet-and-bye.elv`:

```
name = $args[0]
echo Hello, $name!
echo Bye, $name!
```

We can run it like this:

```elvish-transcript
~> elvish greet-and-byte.elv Jane
Hello, Jane!
Bye, Jane!
~> elvish greet-and-byte.elv John
Hello, John!
Bye, John!
```

# Output capture and multiple values

Environment variables are not the only way to learn about a computer system;
we can also gain more information by invoking commands. The `uname` command
tells you which operation system the computer is running; for instance, if you
are running Linux, it prints `Linux` (unsurprisingly):

```elvish-transcript
~> uname
Linux
```

(If you are running macOS, `uname` will print `Darwin`, the [open-source
core](https://en.wikipedia.org/wiki/Darwin_(operating_system)) of macOS.)

Let's try to integrate this information into our "hello" message. The Elvish
command-line allows us to run multiple commands in a batch, as long as they
are separated by semicolons. We can build the message by running multiple
commands, using `uname` for the OS part:

```elvish-transcript
~> echo Hello, $E:USER, ; uname ; echo user!
Hello, xiaq,
Linux
user!
```

This has the undesirable effect that "Linux" appears on its own line. Instead
of running this command directly, we can first **capture** its output in a
variable:

```elvish-transcript
~> os = (uname)
~> echo Hello, $E:USER, $os user!
Hello, elf, Linux user!
```

You can also use the output capture construct directly as an argument to
`echo`, without storing the result in a variable first:

```elvish-transcript
~> echo Hello, $E:USER, (uname) user!
Hello, elf, Linux user!
```

## More arithmetics

You can use output captures to construct do complex arithmetic involving more
than one operation:

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

