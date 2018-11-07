<!-- toc number-sections -->

**Work in progress.**

This tutorial introduces the fundamentals of shell programming with Elvish.
Familiarity with other shells or programming languages is useful but not
required.


# Commands

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
quotes** tells Elvish that the text inside it is a single argument; they are
not part of the argument.

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

TODO: ttyshot

This will give you the last command you have run. However, it may have been a
while when you have last run the `randint` command, and this will not give you
what you need. You can either continue pressing <span class="key">Up</span>
until you find the command, or you you can give Elvish a hint by typing some
characters from the command line you want, e.g. `ra`, before pressing <span
class="key">Up</span>:

TODO: ttyshot

Another way to rerun commands is saving them in a **script**, which is simply
a text file containing the commands you want to run. Using your favorite text
editor, save the command to `dice.elv` under your home directory.
Alternatively, you can use `cat` to write the script:

```elvish-transcript
~> cat > dice.elv
# Type the following in the terminal
randint 1 7
# Now press Enter followed by Ctrl-D
```

After saving the script, you can run it with:

```elvish-transcript
~> elvish dice.elv
▶ 4
```

Since the above command runs `elvish` explicitly, it works in other shells as
well, not just from Elvish itself.


# Variables

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

**To be continued...**

<!--

# Output capture

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

