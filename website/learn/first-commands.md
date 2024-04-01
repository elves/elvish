<!-- toc number-sections -->

This article is part of the *Beginner's Guide to Elvish* series:

-   **Your first Elvish commands**

-   [Arguments and outputs](arguments-and-outputs.html)

-   [Variables and loops](variables-and-loops.html)

-   [Pipelines and IO](pipelines-and-io.html)

-   [Value types](value-types.html)

-   [Organizing and reusing code](organizing-and-reusing-code.html)

# Series introduction

Welcome to Elvish!

This series of articles will teach you the basics of Elvish, covering both its
programming language and interactive features. We will be working through
practical examples and see how Elvish can help you unlock your productivity at
the command line.

You don't need to be experienced in other shells, but you need to be familiar
with basic operating system concepts like file systems and programs.

Before you start, make sure that you have [installed Elvish](../get/). After
that, type `elvish` in the terminal and press <kbd>Enter</kbd> to start Elvish
(`$` represents the existing shell's prompt and is not part of what you type):

```sh-transcript Terminal
$ elvish
```

All the examples below assume that you have started Elvish. Alternatively, you
can [use Elvish as your default shell](../get/default-shell.html) and have
Elvish launched by default.

# Hello World!

Let's begin with the classical ritual of printing
["Hello, World!"](https://en.wikipedia.org/wiki/%22Hello,_World!%22_program):

```elvish-transcript Terminal - elvish
~> echo Hello, world!
Hello, world!
```

You can follow this example (and many others) by typing the code and pressing
<kbd>Enter</kbd>. The `~>` part is Elvish's prompt; you don't need to type it.

Elvish code often follows a "**command** + **arguments**" structure. In this
case, the command is [`echo`](../ref/builtin.html#echo), whose job is to just
print out its arguments.

# Builtin commands

The `echo` command does a very simple job, but that's just one of many commands
Elvish supports. Running other commands follows a similar "command + arguments"
structure.

For instance, the [`randint`](../ref/builtin.html#randint) command takes two
arguments *a* and *b* and generates a random integer from the set {*a*, *a*+1,
..., *b*-1}. You can use it as a digital dice:

```elvish-transcript Terminal - elvish
~> randint 1 7
▶ (num 3)
```

(We'll get to the `▶` notation later in this article, and the `(num 3)` notation
in [Value types](value-types.html).)

Arithmetic operations are also commands. Like the `echo` and `randint` commands
we have seen, the command names comes first, making the syntax a bit different
from common mathematical notations (sometimes known as
[Polish notation](https://en.wikipedia.org/wiki/Polish_notation)):

```elvish-transcript Terminal - elvish
~> + 2 10 # addition
▶ (num 12)
~> * 2 10 # multiplication
▶ (num 20)
```

These commands come with Elvish and are thus called **builtin commands**. The
reference page for the [builtin module](../ref/builtin.html) contains all the
builtin commands you can use directly.

## Commands in modules

Some builtin commands live in **modules** that you need to **import** first. For
example, mathematical operations such as the power function is provided by the
[`math` module](../ref/math.html):

```elvish-transcript Terminal - elvish
~> use math # import the "math" module
~> math:pow 2 10 # raise 2 to the power of 10
▶ (num 1024)
```

We'll learn more about modules in
[Organizing and reusing code](organizing-and-reusing-code.html).

## Comment syntax

You may have noticed that we have annotated some of our commands above with
texts like `# texts`:

```elvish-transcript Terminal - elvish
~> + 2 10 # addition
▶ (num 12)
```

The `#` character marks the rest of the line as a **comment**, which is ignored
by Elvish. When typing out the examples, you can include the comments or leave
them out.

# External commands

While Elvish provides a lot of useful functionalities as builtin commands, it
can't do everything. This is where **external commands** come in, which are
separate programs installed on your machine.

Many useful programs come in the form of external commands, and there is no
limit on what they can do. Here are just a few examples:

-   [Git](https://git-scm.com) provides the `git` command to manage code
    repositories

-   [Pandoc](http://pandoc.org) provides the `pandoc` command to convert
    document formats

-   [GraphicsMagick](http://www.graphicsmagick.org) provides the `gm` command,
    and [ImageImagick](https://www.imagemagick.org/script/index.php) provides
    the `magick` command to process images

-   [FFmpeg](http://ffmpeg.org) provides the `ffmpeg` command to process and
    videos

-   [Curl](https://curl.se) provides the `curl` command to make HTTP requests

-   [Nmap](https://nmap.org) provides the `nmap` command to test the security of
    websites

-   [Tcpdump](http://www.tcpdump.org) provides the `tcpdump` command to analyze
    network traffic

-   Elvish itself can be used as an external command, as `elvish`

-   Your operating system comes with many of external commands pre-installed.
    For example, Unix systems provide [`ls`](https://en.wikipedia.org/wiki/Ls)
    for listing files, and [`cat`](https://en.wikipedia.org/wiki/Cat_(Unix)) for
    showing files. Both Unix systems and Windows provide
    [`ping`](https://en.wikipedia.org/wiki/Ping_(networking_utility)) for
    testing network connectivity.

These example are all command-line programs, but even graphical programs often
provide command-line interfaces, which can give you access to advanced
configuration options. For example, you can invoke
[Chromium](https://www.chromium.org/developers/how-tos/run-chromium-with-flags/)
like `chromium --js-flags=...` to customize internal JavaScript options (the
exact command name depending on the operating system).

## A concrete example

Much of the power of shells comes from the ease of running external commands,
and Elvish is no exception. Here is an example of how you can download Elvish,
using a combination of external commands:

-   `curl` to download files over HTTP

-   `tar` to unpack archive files

-   `shasum` (on Unix systems) or `certutil` (on Windows) to calculate the
    [SHA-256 checksum](https://en.wikipedia.org/wiki/SHA-2) of files

```elvish-transcript Terminal - elvish
~> curl -s -o elvish-HEAD.tar.gz https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz
~> curl -s https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz.sha256sum
93b206f7a5b7f807f6b2b2b99dd4074ed678620541f6e9742148fede0a5fefdb  elvish-HEAD.tar.gz
~> shasum -a 256 elvish-HEAD.tar.gz # On a Unix system
93b206f7a5b7f807f6b2b2b99dd4074ed678620541f6e9742148fede0a5fefdb  elvish-HEAD.tar.gz
~> certutil -hashfile elvish-HEAD.tar.gz SHA256 # On Windows
93b206f7a5b7f807f6b2b2b99dd4074ed678620541f6e9742148fede0a5fefdb
~> tar -xvf elvish-HEAD.tar.gz
x elvish
~> ./elvish
```

(Note: To install Elvish, you're recommended to use the script generated on the
[Get Elvish](../get/) page instead. This example is for illustration, and
assumes your OS to be Linux and CPU to be Intel 64-bit. We will see how to avoid
making these assumptions in [Variables and loops](variables-and-loops.html).)

Running external commands follow the same "command + arguments" structure. For
example, in the first `curl` command, the arguments are `-s`, `-o`,
`elvish-HEAD.tar.gz` and `https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz`
respectively.

External commands often assign special meanings to arguments starting with `-`,
sometimes called **flags**. In the case of `curl`, the arguments work as
follows:

-   `-s` means "silent mode". You can try leaving this argument out to see more
    output from `curl`.

-   `-o`, along with the next argument `elvish-HEAD.tar.gz`, means that `curl`
    should output the result to `elvish-HEAD.tar.gz`.

-   `https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz` is the URL to request.

The second `curl` command leaves out the `-o` and its next argument, directing
`curl` to write the result directly as output. This allows us to see the
checksum directly in the terminal.

You can find out what flags each external command accepts in their respective
manual. You'll also find commands that accept flags starting with `--`, and on
Windows, starting with `/`.

# The working directory

So far, we have run all the commands from the prompt `~>`. The `>` part only
functions as a visual clue, but the `~` part indicates the
[**working directory**](https://en.wikipedia.org/wiki/Working_directory) of
Elvish. In this case, the special `~` indicates your home directory, which can
be /home/*username*, /Users/*username*, or C:\\Users\\*username*, depending on
the operating system.

Commands that work with files are affected by the working directory, especially
when you use a
[**relative path**](https://en.wikipedia.org/wiki/Path_(computing)#Absolute_and_relative_paths)
like `foo` (as opposed to an absolute path like `/etc/foo` or `C:\foo`). Let's
look at our last example more carefully:

-   The first `curl` command writes `elvish-HEAD.tar.gz` to the working
    directory.

-   The `shasum` or `certutil` command reads `elvish-HEAD.tar.gz` from the
    working directory.

-   The `tar` command reads `elvish-HEAD.tar.gz` from the working directory, and
    writes `elvish` to the working directory.

You can change the working directory with the [`cd`](../ref/builtin.html#cd)
command, and all these commands will work with files in that directory instead.
For example, let's create a `tmp` directory under the home directory and use
that as our working directory instead:

```elvish-transcript Terminal - elvish
~> mkdir tmp # on Unix
~> md tmp    # on Windows
~> cd tmp
~/tmp> # Now run more commands...
```

You'll notice that Elvish's prompt has changed to `~/tmp` (or `~\tmp` on
Windows) to reflect the new working directory.

Our examples will continue to use the home directory for simplicity, but it's
recommended that you try them in a separate directory and keep the home
directory clean.

## Directory history

As you work with different directories on your filesystem, you'll find yourself
using `cd` very frequently to jump back and forth between directories, and this
can become quite a chore.

Elvish actually remembers all the directories you have been to, and you can
access that history with the **location mode** by pressing <kbd>Ctrl-L</kbd>:

```ttyshot Terminal - elvish
learn/first-commands/location-mode
```

Press <kbd>▲</kbd> and <kbd>▼</kbd> to select a directory, and press
<kbd>Enter</kbd> to use it as your working directory.

You can also just type part of the path to narrow down the list. For example, to
only see paths that contain `elv`:

```ttyshot Terminal - elvish
learn/first-commands/location-mode-elv
```

# Editing the code interactively

If you have tried all the examples on your computer, it's possible that you have
made some typos. (If you haven't made any, pause and appreciate your typing
skills.)

You are likely already familiar with these keys:

-   <kbd>◀︎</kbd> and <kbd>▶︎</kbd> move the cursor one character at a time.

-   <kbd>Alt-◀︎</kbd> and <kbd>Alt-▶︎</kbd> move the cursor one word at a time.

-   <kbd>Home</kbd> and <kbd>End</kbd> move the cursor to the start or end of
    the line.

-   <kbd>Backspace</kbd> deletes the character to the left of the cursor.

-   <kbd>Delete</kbd> deletes the character on the cursor (or to its right if
    your cursor is an I-beam).

Elvish also supports some keys found in traditional shells:

-   <kbd>Ctrl-W</kbd> deletes the word to the left of the cursor.

-   <kbd>Ctrl-U</kbd> deletes the entire line, up to the cursor.

(See [`readline-binding`](../ref/readline-binding.html) if you'd like to use
more readline-style bindings.)

# Conclusion

Being able to run all sorts of commands easily is one of the greatest strengths
of shells. When using Elvish interactively, most of your interactions will
consist of simple invocations of various commands.

In this part, we learned the basics of running builtin and external commands and
how they can be affected by the working directory. We also learned how to change
the working directory quickly and how to edit your command.

We are now ready for the next part,
[Arguments and outputs](arguments-and-outputs.html).
