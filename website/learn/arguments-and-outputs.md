<!-- toc number-sections -->

This article is part of the *Beginner's Guide to Elvish* series:

-   [Your first Elvish commands](first-commands.html)

-   **Arguments and outputs**

-   [Variables and loops](variables-and-loops.html)

-   [Pipelines and IO](pipelines-and-io.html)

-   [Value types](value-types.html)

-   [Organizing and reusing code](organizing-and-reusing-code.html)

# Arguments and quoting

Let's take a closer look at some of the commands we've run:

```elvish-transcript Terminal - elvish
~> + 2 10
▶ (num 12)
~> echo Hello, world!
Hello, world!
```

In the first command, `+` is given two arguments, `2` and `10`, separated by
spaces -- so far so good.

In the second command, `echo` is given `Hello, world!`. Following how we read
the first command, this should also be *two* arguments. But this still achieved
what we want -- printing out `Hello, World!` -- but how?

As it turns out, what's happening with this simple command is actually not so
simple:

1.  Elvish recognizes `Hello, World!` as two arguments and passes them to
    `echo`.

2.  The `echo` command receives two arguments and it joins them with a space
    before printing.

The consequence of this process can be better observed when you'd like to print
multiple spaces:

```elvish-transcript Terminal - elvish
~> echo Hello,   world!
Hello, world!
```

Although we've used three spaces, all `echo` receives is two arguments `Hello,`
and `world!`, so all it can do is joining them back with a single space.

## Quoting

In order to pass spaces to `echo`, we can pass a **quoted** argument:

```elvish-transcript Terminal - elvish
~> echo 'Hello,   world!'
Hello,  world!
```

A pair of single quotes delimits a **quoted string**, and tells Elvish that the
text inside it is a single argument. In this case, `echo` only sees one argument
containing <code>Hello,&nbsp;&nbsp;&nbsp;world</code> (with three spaces).

You can use quoted arguments wherever you can use an unquoted argument. In fact,
you can even quote the command name itself. This applies to all the commands we
have seen so far:

```elvish-transcript Terminal - elvish
~> '+' '2' '10'
▶ (num 12)
~> 'randint' '1' '7'
▶ (num 3)
```

However, writing commands like this is unnecessary and unreadable. It's better
to reserve quoting to situations when it's necessary, for example when the
argument you'd like to pass has spaces.

## Metacharacters

Another common reason to use quoting is when your argument includes
**metacharacters**, characters with special meaning to Elvish. Some examples of
metacharacters are `#`, which introduces a comment, and `(` and `)`, which we'll
encounter soon. Quoting them stops Elvish from treating them as metacharacters:

```elvish-transcript Terminal - elvish
~> echo '(1) Rule #1'
(1) Rule #1
```

We will learn more metacharacters as we learn more of Elvish's syntax. As a rule
of thumb, punctuation marks in the
[ASCII range](https://en.wikipedia.org/wiki/ASCII#Character_set) tend to be
metacharacters, except those commonly found in file paths, like `_`, `-`, `.`,
`/` and `\`.

## Double quotes

You can also create quoted strings with double quotes, like `"foo"` rather than
`'foo'`. It's useful when the string itself contains single quotes:

```elvish-transcript Terminal - elvish
~> echo "It's a wonderful world!"
It's a wonderful world!
```

Another difference is that double quotes allow you to write special characters
using **escape sequences** that start with `\`. For example, `\n` inside double
quotes represents a newline:

```elvish-transcript Terminal - elvish
~> echo "old pond\nfrog leaps in\nwater's sound"
old pond
frog leaps in
water's sound
```

You can read more about
[single-quoted strings](../ref/language.html#single-quoted-string) and
[double-quoted strings](../ref/language.html#double-quoted-string) in the
language reference page.

# Working with command outputs

Let's go back to our arithmetic examples:

```elvish-transcript Terminal - elvish
~> + 2 10 # addition
▶ (num 12)
~> * 2 10 # multiplication
▶ (num 20)
```

What if we want to calculate something that involves multiple operations, like
`3 * (2 + 10)`? Of course, we can simply run multiple commands:

```elvish-transcript Terminal - elvish
~> + 2 10
▶ (num 12)
~> * 3 12
▶ (num 36)
```

But this makes Elvish a really clumsy calculator! Instead, we can **capture**
the output of the `+` command by placing it inside `()`, and use it as an
argument to the `*` command:

```elvish-transcript Terminal - elvish
~> * (+ 2 10) 3
▶ (num 36)
```

Now is also the time to explain the `▶` notation. It indicates that the output
is a **value** in Elvish's data type system. In this case, `(num ...)` is a
**typed number**, although for the purpose of passing to `*`, `(num 12)` and
`12` work identically. We will learn more in [Value types](value-types.html).

## Concatenating results

Other than using the output of a command as an argument on its own, you can also
concatenate it to something else to build a bigger argument. For example, if we
want to add some message explaining what the result is:

```elvish-transcript Terminal - elvish
~> echo 'And the result is... '(* (+ 2 10) 3)
And the result is... 36
```

(For the purpose of string concatenation, `(num 36)` becomes just `36`.)

When the output of a command doesn't start with `▶`, it indicates that the
output is a stream of *bytes*. For example, `echo` produces bytes:

```elvish-transcript Terminal - elvish
~> echo Hello!
Hello!
```

Among Elvish's builtin commands, some output values, while some output bytes. On
the other hand, external commands can only output bytes; they don't have direct
access to Elvish's system of data types.

Let's finish this section by augmenting our "Hello, World!" example with the
output of a useful external commands:

```elvish-transcript Terminal - elvish
~> echo 'Hello World! My name is: '(whoami)
Hello World! My name is: elf
```

# Command history

We have now worked with quite a few commands, some more simple, some more
complex. Inevitably, you'll want to run some commands that are either the same
or similar to something you have run in the past. This is where Elvish's
**command history** feature is useful.

For example, if we rolled the dice once with `randint 1 7` and want to roll it
again, we can press <kbd>▲</kbd>:

```ttyshot Terminal - elvish
learn/arguments-and-outputs/command-history-up
```

We are now in a **history walking** mode. The basic operations work as follows:

-   Press <kbd>▲</kbd> to go back further, or <kbd>▼</kbd> to go forward.

-   If you can't find your command, press <kbd>Esc</kbd> to exit this mode
    cleanly.

-   Pressing any other key accepts the current entry and do what the key usually
    does. For example, simply pressing <kbd>Enter</kbd> will accept the entry
    and run it. If you want to edit the command before running it, just start
    editing by pressing <kbd>Backspace</kbd>, <kbd>Ctrl-W</kbd>, and so on.

Walking the history like this is the best option if the command is recent. To
find a more distant command, you can use the **history listing** mode instead by
pressing <kbd>Ctrl-R</kbd>:

```ttyshot Terminal - elvish
learn/arguments-and-outputs/command-history-listing
```

This mode works very similarly to **location mode** we saw in
[Your first Elvish commands](first-commands.html). You can use <kbd>▲</kbd> and
<kbd>▼</kbd> to select an item, and press <kbd>Enter</kbd> to insert to it.
Similarly, you can type part of the command to narrow down the list:

```ttyshot Terminal - elvish
learn/arguments-and-outputs/command-history-listing-narrowed
```

# Conclusion

In this part, we dived into the inner workings of arguments and quoting, and
learned how to capture and use the output of commands, and the distinction
between *value output* and *byte output*. We also learned how to recall command
history. These skills will help you build and use more complex commands and use
them with ease.

Let's now move on to the next part,
[Variables and loops](variables-and-loops.html).
