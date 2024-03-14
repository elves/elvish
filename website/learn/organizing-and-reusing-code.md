<!-- toc number-sections -->

This article is part of the *Beginner's Guide to Elvish* series:

-   [Your first Elvish commands](first-commands.html)

-   [Arguments and outputs](arguments-and-outputs.html)

-   [Variables and loops](variables-and-loops.html)

-   [Pipelines and IO](pipelines-and-io.html)

-   [Value types](value-types.html)

-   **Organizing and reusing code**

# Scripts

So far, we have run all our Elvish commands directly at Elvish's prompt. This
can be quite convenient: you don't have to open an editor or compile your code,
just open up your terminal, type something, and see the results. This is true of
shells in general, but Elvish also gives you a powerful programming language
capable of processing complex data.

Still, from time to time, we'll need to organize our code a bit more formally,
for example when we need to send it to a different machine or someone else. We
can achieve that by simply putting our code in a file. As an example, let's open
an editor and type our first very "Hello, world!" program (albeit with proper
quoting):

```elvish hello.elv
echo 'Hello, world!'
```

After saving the file as `hello.elv`, you can run it like:

```elvish-transcript Terminal - elvish
~> elvish hello.elv
Hello, world!
```

Such file are usually called **scripts**. Elvish scripts use the extension
`.elv` by convention; this is not a requirement from Elvish itself, but naming
the file with an `.elv` extension communicates to other people that this is an
Elvish script, and makes it easier for your editor to detect the file's type.

# Functions

We have seen a lot of commands, both builtin and external. Elvish also gives you
the ability to define your own commands.

For example, in
[Variables and loops](variables-and-loops.html#for-loops-and-lists), we used
`gm` to convert JPG files to AVIF files. If we need to perform this conversion
frequently, we can define a **function** that takes the name of a JPG file:

```elvish-transcript Terminal - elvish
~> use str
~> fn jpg-to-avif {|jpg| gm convert $jpg (str:trim-suffix $jpg .jpg).avif }
```

After that, you can use `jpg-to-avif` like any other command:

```elvish-transcript Terminal - elvish
~> jpg-to-avif unicorn.jpg
~> jpg-to-avif banana.jpg
```

The [`fn`](../ref/language.html#fn) command defines a function. Here, the
function to define is called `jpg-to-avif`, and the part surrounded by `{` and
`}` is its **body**.

The body starts with `|jpg|`, and it specifies that the function takes a single
argument, which becomes a variable `$jpg`. The `|` that surrounds the argument
is the same character we use for pipes, but has a different meaning in this
context.

In Elvish, we also call builtin commands "builtin functions", but external
commands are not called functions.

## rc.elv

Defining a function at the prompt only makes it available for the current
session. If you open a new terminal, you'll have to define it again.

To define the function automatically, you can put it in a special `rc.elv`
script, which is evaluated before every interactive Elvish session. The path
[depends on the system](https://elv.sh/ref/command.html#rc-file), but by
default, it's `~/.config/elvish/rc.elv` on Unix systems and
`%RoamingAppData%\elvish\rc.elv` on Windows.

We can add the following to `rc.elv` (create it if it doesn't exist yet):

```elvish rc.elv
use str
fn jpg-to-avif {|jpg| gm convert $jpg (str:trim-suffix $jpg .jpg).avif }
```

## Functions as aliases

The `jpg-to-avif` function does something relatively complex, but even very
simple functions can be useful. For example, it's a good idea to upgrade your
system packages every day, so you may want to define a dedicated function for it
and put that in `rc.elv`:

```elvish rc.elv
fn up { brew upgrade }
```

(We are using the [Homebrew](https://brew.sh) package manager as an example;
change the exact command according to the package manager your system uses.)

Note that we have omitted the list of arguments because this function doesn't
take any. It's equivalent to:

```elvish rc.elv
fn up {|| brew upgrade }
```

(Due to some quirks in Elvish's syntax, you have to follow the `{` with a
whitespace character, such as a space or a newline.)

Even though our definition of `up` is quite simple, it can still save us a lot
of keystrokes if we upgrade our system frequently. This kind of simple functions
are sometimes called **aliases**. Some shells have aliases as a distinct concept
from functions, but in Elvish they are the same.

Another popular alias among Unix users is `ll` for `ls -l`. We can define it
like this:

```elvish rc.elv
fn ll { ls -l }
```

Our definition has a defect, however. The `ls` command has two modes of
operations:

-   If you run it without any arguments, it lists the current directory.

-   Alternatively, you can also give it any numbers of files you are interested
    in, like this:

    ```elvish-transcript Terminal - elvish
    ~> ls -l foo bar
    [ information about foo and bar ]
    ```

Our `ll` function only supports the former mode. To fix that, we can let it
accept any number of arguments too:

```elvish rc.elv
fn ll {|@a| ls -l $@a }
```

The `@` in `@a` causes elvish to collect an arbitrary number of arguments into
`$a` as a list. We then use `$@a` to "expand" it back into individual arguments.

## Functions as arguments

The function body syntax is not restricted to function definitions. Elvish has
**first-class functions**, meaning that you can use functions as arguments to
other commands too. (See
[Wikipedia](https://en.wikipedia.org/wiki/First-class_function) for the general
concept).

We have actually seen a few of those. For example, the `each` command takes a
function and run it for each of the inputs.

```elvish-transcript Terminal - elvish
~> put 1 2 3 | each {|n| * $n 2}
▶ (num 2)
▶ (num 4)
▶ (num 6)
```

A slightly more subtle occurrence is the body of `for` and `if` commands, which
look like `{ commands }`. These are in fact functions that don't take any
arguments.

# More on scripts

Like functions, scripts can also take arguments. Let's try running `hello.elv`
with some arguments:

```elvish-transcript Terminal - elvish
~> elvish hello.elv foo bar
Hello, world!
```

As you can see, this doesn't change the behavior of our script. That is because
we aren't actually using the arguments: unlike functions which declare their
arguments inside `||`, arguments to scripts are available implicitly as a list
in `$args`.

Let's make our scripts treat each argument as someone to say hello to, falling
back to `world` if there are no arguments:

```elvish hello.elv
if (== 0 (count $args)) {
  echo 'Hello, world!'
} else {
  for who $args {
    echo 'Hello, '$who'!'
  }
}
```

Here, the [`==`](../ref/builtin.html#num-eq) command compares two numbers, and
the [`count`](../ref/builtin.html#count) command counts the number of elements
in a list.

We can check that the new `hello.elv` works as intended:

```elvish-transcript Terminal - elvish
~> elvish hello.elv
Hello, world!
~> elvish hello.elv Julius Augustus
Hello, Julius!
Hello, Augustus!
```

One important thing to keep in mind is that the command `elvish hello.elv`
behaves like an external command. Even though it's the same program as the
Elvish you run it from, it's a separate
[process](https://en.wikipedia.org/wiki/Process_(computing)). You can still use
Elvish's system of values within the `hello.elv` script itself, but it can't
communicate with the "outer world" using Elvish values, only string arguments
and byte IO.

# Modules

In our past exampls, we have often use the following pattern to access
additional commands provided by Elvish:

```elvish-transcript Terminal - elvish
~> use str                    # ①
~> str:trim-suffix a.jpg .jpg # ②
▶ a
```

1.  This command **imports** a **module** to make it available for use.

    In this case, the module is `str`, and as its abbreviated name suggests, it
    provides commands for working with strings.

2.  To use a command that lives inside a module, we need to prefix it with the
    module name plus a colon `:`. The technical way to put this is that all the
    commands in a module lives in a separate **namespace** (see
    [Wikipedia](https://en.wikipedia.org/wiki/Namespace) for the general
    concept).

    (You'll sometimes see the colon treated as part of the module name itself,
    to make it clear that we are referring to a module; we may say either "the
    `str` module" or just `str:`.)

Elvish has many more builtin modules, and you can see them in the
[reference](../ref/) section.

Organizing commands into separate modules makes them easier to discover, and the
separate namespaces prevent
[name collisions](https://en.wikipedia.org/wiki/Name_collision). For example,
there is both a [`str:replace`](../ref/str.html#str:replace) command and a
[`re:replace`](../ref/re.html#re:replace) command: the former replaces simple
literal strings, the latter works with regular expressions.

## Defining and using new modules

Just like how you can define your own functions, you can also define your own
modules. Do this by placing a file under a module search directory: like
`rc.elv`, the path of the directory
[depends on the system](../ref/command.html#module-search-directories), but by
default, `~/.config/elvish/lib` works on Unix systems and
`%RoamingAppData%\elvish\lib` works on Windows.

For example, let's collect the `jpg-to-avif` commands into a `img` module, since
we may have more of them in future. Create `img.elv` under a module search
directory:

```elvish img.elv
use str
fn jpg-to-avif {|jpg| gm convert $jpg (str:trim-suffix $jpg .jpg).avif }
```

After that, you can use it like this:

```elvish-transcript Terminal - elvish
~> use img
~> img:jpg-to-avif unicorn.jpg
```

Notice that when using a module with `use`, we omit the `.elv` file extension.

## Modules in subdirectories

You don't have to put modules directly under a module search directory; you can
also store it in a subdirectory. For example, let's collect our `img` module and
other modules we have into a `myutils` directory:

```elvish myutils/img.elv
use str
fn jpg-to-avif {|jpg| gm convert $jpg (str:trim-suffix $jpg .jpg).avif }
```

Then you would use it like this:

```elvish-transcript Terminal - elvish
~> use myutils/img
~> img:jpg-to-avif unicorn.jpg
```

Notice that the `use` command takes the full path to the module (relative to the
module search directory), but after that, we'll just use the last part to access
it.

# Conclusion

In this part, we've covered scripts, functions and modules, important mechanisms
that allow you to organize code and reuse them. We've also seen how Elvish's
support for first-class functions enables commands like `each`, and how Elvish's
namespacing mechanism in the module system prevents name conflicts.

# Series conclusion

Congratulations for finishing the *Beginner's Guide to Elvish* series! We
haven't covered everything, but what we have learned should give you a solid
basis to build upon, and already allow you to be productive in your daily
workflows.

You can read more articles in the [learn](./) section, or go directly to
[reference manuals](../ref/) (in particular the
[language specification](../ref/language.html)). The latter can be a bit dense,
but they will give you a complete understanding of how Elvish works, and you
should be ready to read them after going through this series.

Have fun with Elvish!
