<!-- toc number-sections -->

This article is part of the *Beginner's Guide to Elvish* series:

-   [Your first Elvish commands](first-commands.html)

-   [Arguments and outputs](arguments-and-outputs.html)

-   [Variables and loops](variables-and-loops.html)

-   **Pipelines and IO**

-   [Value types](value-types.html)

-   [Organizing and reusing code](organizing-and-reusing-code.html)

# Introduction

In the previous parts, we have seen how you can combine commands with *output
capture*: you can use the outputs of a command as the arguments to another
command. You can also store them in variables or concatenate them with something
else.

Elvish and other shell languages have another powerful mechanism for combining
commands called **pipelines**. But before that, let's examine how input and
output of commands work.

# IO redirection

So far, we have worked with commands that take *arguments* and write *outputs*.
Some commands also take **inputs**.

As an example, the [`grep`](https://en.wikipedia.org/wiki/Grep) command on Unix
systems and the
[`findstr`](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/findstr)
command on Windows reads lines of text, matches each line with a pattern, and
writes matched lines as output.

This may sound quite complicated, so it's best illustrated with an example.
Let's save the following content in a file called `input.txt`:

```
User A: I use Bash
User B: I use Elvish
User C: I use Zsh
User D: I use Elvish
```

After that, let's run the following command to find the lines that contain
`Elvish`:

```elvish-transcript Terminal - elvish
~> grep Elvish < input.txt    # on Unix systems
~> findstr Elvish < input.txt # on Windows
User B: I use Elvish
User D: I use Elvish
```

The `< input.txt` syntax is **input redirection** and tells Elvish to use a file
as a command's input. Here, we can see `grep` or `findstr` correctly finds the
two lines in the file that contains `Elvish`.

You can also perform **output redirection** with `>`:

```elvish-transcript Terminal - elvish
~> grep Elvish < input.txt > output.txt    # on Unix systems
~> findstr Elvish < input.txt > output.txt # on Windows
```

This command won't have any outputs on the terminal, but if you open
`output.txt`, it should contain the same two lines.

Input and output are known collectively as **IO**. IO redirection is useful for
processing data stored in files and saving results in files.

## Typing inputs from the terminal

We can also run commands like `grep` or `findstr` without redirecting their
input:

```elvish-transcript Terminal - elvish
~> grep Elvish > output.txt    # on Unix systems
~> findstr Elvish > output.txt # on Windows
```

(You can also try running the command without the output redirection, but the
interleaving of input and output lines can make things a bit confusing.)

After running the command, you'll find your cursor on an empty line without
Elvish's `~>` prompt. This is because the `grep` or `findstr` command is still
running and waiting for your input. Try typing the same lines that we saved in a
file before:

```
User A: I use Bash
User B: I use Elvish
User C: I use Zsh
User D: I use Elvish
```

You also need to indicate that you have finished typing:

1.  Make sure that you are on an empty line. If not, press <kbd>Enter</kbd>.

2.  The next step depends on the operating system:

    -   On Unix systems, press <kbd>Ctrl-D</kbd>.

    -   On Windows, press <kbd>Ctrl-Z</kbd> and another <kbd>Enter</kbd>.

If you open up `output.txt`, it should have the two lines containing `Elvish`.

Typing inputs directly from the terminal can be quite handy when you have a
small amount of input; when you have a lot of input, it's best to store them in
a file and use input redirection instead.

# Traditional byte pipelines

We have seen how we can make `grep` or `findstr` work with files, but another
more powerful way to use them is by feeding it the *output* of another command.

Here is an example. We have used `curl` to retrieve <https://dl.elv.sh/INDEX>,
an index of files provided by <https://dl.elv.sh>. If we're only interested in
the `HEAD` version, we can filter the output of `curl` by connecting it to
`grep` or `findstr` using `|`, a **pipe**:

```elvish-transcript Terminal - elvish
~> curl -s https://dl.elv.sh/INDEX | grep HEAD # Unix
~> curl -s https://dl.elv.sh/INDEX | findstr HEAD # Windows
https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz
https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz.sha256sum
https://dl.elv.sh/darwin-arm64/elvish-HEAD.tar.gz
https://dl.elv.sh/darwin-arm64/elvish-HEAD.tar.gz.sha256sum
...
```

Like real-world pipelines, you can extend pipelines with more pipes. If we're
also only interested in Linux builds, we can add *another* `grep` or `findstr`:

```elvish-transcript Terminal - elvish
~> curl -s https://dl.elv.sh/INDEX | grep HEAD | grep linux # Unix
~> curl -s https://dl.elv.sh/INDEX | findstr HEAD | findstr linux # Windows
https://dl.elv.sh/linux-386/elvish-HEAD.tar.gz
https://dl.elv.sh/linux-386/elvish-HEAD.tar.gz.sha256sum
https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz
https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz.sha256sum
...
```

## Unix's text processing toolkit

The `grep`/`findstr` command is just one example of commands that work well in
pipelines. In particular, Unix systems, where the idea of pipelines
[originate](https://en.wikipedia.org/wiki/Pipeline_(Unix)), provide many more
commands designed to work in pipelines.

Suppose you are running
[Docker](https://en.wikipedia.org/wiki/Docker_(software)) containers and need to
read the logs of a container. You can use the `docker` command like this:

```elvish-transcript Terminal - elvish
~> docker logs container-name
...
```

However, this gives you *all* the logs, which you probably don't need. We can
use `grep` to only show lines that match a pattern:

```elvish-transcript Terminal - elvish
~> docker logs container-name | grep error
...
```

Or use [`head`](https://en.wikipedia.org/wiki/Head_(Unix)) or
[`tail`](https://en.wikipedia.org/wiki/Tail_(Unix)) to only show the first or
last N lines:

```elvish-transcript Terminal - elvish
~> docker logs container-name | head -n 20 # read first 20 lines
...
~> docker logs container-name | tail -n 20 # read last 20 lines
...
```

Here's a more elaborate example: the following pipeline shows the top 10
committers of a Git repo, ranked by their number of commits (the output below is
from the Git repo of [Go](https://go.googlesource.com/go/)):

```elvish-transcript Terminal - elvish
~/src/go> git log --format=%an | sort | uniq -c | sort -nr | head -n 10
7343 Russ Cox
4303 Robert Griesemer
2992 Rob Pike
2613 Ian Lance Taylor
2377 Brad Fitzpatrick
1734 Austin Clements
1704 Matthew Dempsky
1517 Keith Randall
1496 Josh Bleecher Snyder
1346 Bryan C. Mills
```

In fact, whole books have been written on how to craft Unix pipelines, and they
are arguably the original data science toolkit.

Windows has also adopted the pipeline mechanism, but it doesn't come
pre-installed with the commands we have seen in this section. Nonetheless, you
can install them from software sources like [Scoop](https://scoop.sh) and most
Unix pipelines will work on Windows too.

## Limitations of byte pipelines

As powerful as they are, traditional pipelines do have some pretty severe
limitations.

As the name "byte pipeline" suggests, the pipes carry streams of bytes, which
lack an inherent structure. By *convention*, many tools -- like `grep`, `head`
and `tail` -- treat each line as a unit. Some tools also break each line down
into space-separated "fields".

These conventions give byte pipelines some extra mileage, but they are still
limited: you can't easily deal with data that contain newlines or whitespaces
themselves, or data with more complex structure, like multiple levels of lists
or maps.

# Value pipeline

To overcome the limitations of traditional byte pipelines, Elvish offers **value
pipelines**.

We have seen how commands in Elvish can output *values* instead of bytes. For
example, the `str:split` splits a string around a separator, and outputs the
results as string values:

```elvish-transcript Terminal - elvish
~> str:split , friends,Romans,countrymen
▶ friends
▶ Romans
▶ countrymen
```

There are also commands that can take values as *inputs*. The
[`str:join`](../ref/str.html#str:join) command joins multiple strings together,
inserting a separator between each adjacent pairs. We can connect the output of
`str:split` with the input of `str:join` like this:

```elvish-transcript Terminal - elvish
~> str:split , friends,Romans,countrymen | str:join ' '
▶ 'friends Romans countrymen'
```

## Working with both bytes and values

The commands we have worked with either use bytes for both input and output
(like `grep`/`findstr`) or use values for both input and output (like
`str:join`). But that doesn't have to be the case. The
[`from-json`](../ref/builtin.html#from-json) command takes byte inputs, parses
them as [JSON](https://en.wikipedia.org/wiki/JSON), and writes the result as
Elvish values:

```elvish-transcript Terminal - elvish
~> echo '["Julius","Crassus","Pompey"]' | from-json
▶ [Julius Crassus Pompey]
```

(Elvish lists can look a bit like JSON lists. Remember that the leading `▶`
indicates value output, and Elvish list elements are separated by spaces rather
than commas.)

We can also combine pipes and output capture:

```elvish-transcript Terminal - elvish
~> echo '["Julius","Crassus","Pompey"]' | all (from-json)
▶ Julius
▶ Crassus
▶ Pompey
```

It's worth noting that although `from-json` is put in an output capture, it's
still able to read the byte inputs from the pipe. It outputs a single list,
which is then converted by the `all` to all its elements.

There is also a reverse operation of `from-json`:
[`to-json`](../ref/builtin.html#to-json) takes Elvish value as inputs, converts
them to JSON, and writes the result as bytes:

```elvish-transcript Terminal - elvish
~> put [Julius Crassus Pompey] | to-json
["Julius","Crassus","Pompey"]
```

We haven't quite explored the power of value pipelines. We will see more
interesting examples very soon.

# Conclusion

Elvish allows you to manipulate the byte inputs and outputs of commands with
*redirections*, and combine them using *byte pipelines*, a natural and flexible
way to express data processing logic. Some details may differ, but byte I/O
redirection and pipelines work in other shells too.

Elvish infuses pipelines with more power by allowing values to pass through the
pipes. This allows you to express data processing logic that involves data with
more complex structures, although we've only just had a taste of that.

Let's now move on to [Value types](value-types.html) and unlock the full power
of value pipelines.
