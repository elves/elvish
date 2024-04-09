<!-- toc number-sections -->

This article is part of the *Beginner's Guide to Elvish* series:

-   [Your first Elvish commands](first-commands.html)

-   [Arguments and outputs](arguments-and-outputs.html)

-   **Variables and loops**

-   [Pipelines and IO](pipelines-and-io.html)

-   [Value types](value-types.html)

-   [Organizing and reusing code](organizing-and-reusing-code.html)

# Using variables

In [Your first Elvish commands](first-commands.html), we saw
[an example](first-commands.html#a-concrete-example) of how to use a series of
commands to download Elvish. Let's focus on the initial two commands, which
download the archive and show the SHA256 checksum respectively:

```elvish-transcript Terminal - elvish
~> curl -s -o elvish-HEAD.tar.gz https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz
~> curl -s https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz.sha256sum
93b206f7a5b7f807f6b2b2b99dd4074ed678620541f6e9742148fede0a5fefdb  elvish-HEAD.tar.gz
```

This example comes with a catch -- it only works as long as the `linux-amd64`
part actually matches your platform, namely Linux on a
[x86-64 CPU](https://en.wikipedia.org/wiki/X86-64#AMD64). To fix that, instead
of *hardcoding* this string, we need a way to construct it *dynamically* to
actually match your platform.

Turns out that Elvish already has all the information we need, stored inside two
**variables**:

```elvish-transcript Terminal - elvish
~> use platform
~> echo $platform:os
darwin
~> echo $platform:arch
arm64
```

(We'll learn what `use platform` and the colons are about in
[Organizing and reusing code](organizing-and-reusing-code.html)).

The `$` character starts a variable, and tells Elvish to **evaluate** it to the
value stored inside it. In this case, the `$platform:os` variable stores a
string identifying the OS
([`darwin`](https://en.wikipedia.org/wiki/Darwin_(operating_system)) in the
example output), and the `$platform:arch` variable stores a string identifying
the CPU architecture ([`arm64`](https://en.wikipedia.org/wiki/AArch64) in the
example output).

Your output may differ, but at least in the example output, it turns out our
platform doesn't match `linux-amd64` after all. Let's now fix our command by
making use of these variables:

```elvish-transcript Terminal - elvish
~> curl -s -o elvish-HEAD.tar.gz https://dl.elv.sh/$platform:os'-'$platform:arch/elvish-HEAD.tar.gz
~> curl -s https://dl.elv.sh/$platform:os'-'$platform:arch/elvish-HEAD.tar.gz.sha256sum
f1b2e7c149f5104c191bc7c9cd922b87ac73d810ba71c186636d1807e2a5ce95  elvish-HEAD.tar.gz
```

And now our commands work regardless of which platform we are on! In more fancy
terms, our commands are now *portable* across platforms.

Let's recap what is going on:

1.  Elvish sees `$platform:os` and `$platform:arch` and evaluates them to their
    respective values -- in our environment, `darwin` and `arm64` respectively.

2.  Elvish concatenates them to the neighboring strings to form the overall
    argument. The argument for the first `curl` command is
    <https://dl.elv.sh/darwin-arm64/elvish-HEAD.tar.gz>; similarly for the
    second `curl` command, with an extra `.sha256sum` suffix.

3.  The `curl` command then runs with the arguments we have constructed.

(There is still a catch: this example still doesn't work for Windows, because
the archive files for Windows end in `.zip` instead of `.tar.gz`. Once we have
learned conditionals in [Value types](value-types.html), you can come back here
to make this code fully portable.)

## Quoting and syntax highlighting

Notice how we quoted `-` between `$platform:os` and `$platform:arch`. This is
because variable names in Elvish can include `-`, so if we omit it, Elvish will
try to find the variable `$platform:os-`:

```elvish-transcript Terminal - elvish
~> curl -s -o elvish-HEAD.tar.gz https://dl.elv.sh/$platform:os-$platform:arch/elvish-HEAD.tar.gz
Exception: variable $platform:os- not found
```

This introduces us to another reason for quoting strings: when concatenating
literal strings with variables, quoting the literal part can stop Elvish from
treating it as part of the variable name.

Elvish also gives you hints using by **highlighting** different parts of the
code. Let's zoom in on the part around our variables:

```elvish-transcript Terminal - elvish
~> echo $platform:os'-'$platform:arch
darwin-arm64
~> echo $platform:os-$platform:arch
Exception: variable $platform:os- not found
```

In the first correct command, the quoted `'-'` has a distinct color, clearly
standing out from the variables around it. In the second incorrect command, the
unquoted `-` is colored the same as variables, meaning that Elvish will treat it
as part of the variable name.

# Defining new variables

Our commands for downloading Elvish and showing the checksum still has some room
for improvement. Notice how similar the two commands are, in particular the last
argument:

```elvish-transcript Terminal - elvish
~> curl -s -o elvish-HEAD.tar.gz https://dl.elv.sh/$platform:os'-'$platform:arch/elvish-HEAD.tar.gz
~> curl -s https://dl.elv.sh/$platform:os'-'$platform:arch/elvish-HEAD.tar.gz.sha256sum
f1b2e7c149f5104c191bc7c9cd922b87ac73d810ba71c186636d1807e2a5ce95  elvish-HEAD.tar.gz
```

To fix that, we can store the common part in a new variable:

```elvish-transcript Terminal - elvish
~> var archive-url = https://dl.elv.sh/$platform:os'-'$platform:arch/elvish-HEAD.tar.gz
~> curl -s -o elvish-HEAD.tar.gz $archive-url
~> curl -s $archive-url.sha256sum
f1b2e7c149f5104c191bc7c9cd922b87ac73d810ba71c186636d1807e2a5ce95  elvish-HEAD.tar.gz
```

The [`var` command](../ref/language.html#var) defines a new variable called
`archive-url` and gives it an initial value. After that, we can use it like
`$archive-url`.

Notice how we don't use the `$` prefix when defining a variable. This is because
`$` instructs Elvish to evaluate a variable, and we are not doing that when
defining it. However, we may still say that we "define `$archive-url`" as a
shorthand of "define the `archive-url` variable".

# For loops and lists

The ability to use and define variables gives us the flexibility in how we do
*one* thing, but often we find ourselves repeating similar but not entirely
identical tasks.

For example, let's say we have a few `.jpg` files that we would like to convert
into the more efficient [AVIF](https://en.wikipedia.org/wiki/AVIF) format. (If
you'd like to follow this example but don't have spare `.jpg` files lying
around, download some from
[Wikimedia Commons](https://commons.wikimedia.org/w/index.php?search=seashell&title=Special:MediaSearch&go=Go&type=image&filemime=jpeg).)
With the `gm` command provided by
[GraphicsMagick](http://www.graphicsmagick.org), we can convert them one by one:

```elvish-transcript Terminal - elvish
~> gm convert banana.jpg banana.avif
~> gm convert unicorn.jpg unicorn.avif
~> # and so on...
```

There is a better way to do it, though. Like many other programming languages,
Elvish provides *loops* to perform repetitive work:

```elvish-transcript Terminal - elvish
~> use str                                       # ①
~> for jpg [banana.jpg unicorn.jpg] {            # ②
     var avif = (str:trim-suffix $jpg .jpg).avif # ③
     gm convert $jpg $avif                       # ④
   }
```

This is a more complex example, so let's go through it line by line:

1.  `use str` imports the [`str` module](../ref/str.html). We'll learn about
    modules in [Organizing and reusing code](organizing-and-reusing-code.html);
    for now, it suffices to know that this is needed to be able to use
    `str:trim-suffix` below.

2.  The [`for`](../ref/language.html#for) command introduces a **for loop**.

    Let's first focus on `[banana.jpg unicorn.jpg]`: the `[` and `]` delimits a
    **list**, a type of value that consists of multiple **elements**. Here, the
    elements are `banana.jpg` and `unicorn.jpg`, separated by spaces -- just
    like how the arguments to a command are separated by spaces.

    The for loop works as follows: for each element of the list, it defines the
    `jpg` variable to be equal to that element, and runs the code inside `{` and
    `}` (the **body** of the for loop).

    Now for the body itself...

3.  Since the name of the input JPG file is no longer hardcoded, we can no
    longer hardcode the name of the output AVIF file either. Instead, we use
    some string manipulation to derive the output name from the input name - the
    [`str:trim-suffix`](../ref/str.html#str:trim-suffix) commands removes a
    fixed suffix from a string. You can see it in action like this:

    ```elvish-transcript Terminal - elvish
    ~> str:trim-suffix banana.jpg .jpg
    ▶ banana
    ```

    We then concatenate the result with `.avif` to form the output filename, in
    this case `banana.avif`, and store it in the `$avif` variable.

4.  Finally, we use the `gm` command to perform the conversion.

As we can see, the for loop will run the body twice, once with `$foo` equal to
`banana.jpg`, and once with `$foo` equal to `unicorn.jpg`, so this achieves the
same effect as two "manual" invocations `gm` that we set out to improve.

## The strength of loops

In this particular case, we haven't really achieved any improvement -- our new
code is longer and more complex than the two separate `gm` invocations. In fact,
when you only need to repeat a simple task twice or three times, just repeating
it "manually" -- probably with the help of Elvish's command history -- is a
totally valid approach.

The real strength of for loops is when there are many elements, maybe even an
unknown number of them. Let's say we'd like to convert *all* the `.jpg` files to
`.avif` files. With the manual approach you'd have to write as many `gm`
commands as there are files, but with a for loop, just a simple modification is
needed:

```elvish-transcript Terminal - elvish
~> use str
~> for jpg [*.jpg] {                             # ①
     var avif = (str:trim-suffix $jpg .jpg).avif
     gm convert $jpg $avif
   }
```

Here, we have changed the element of the list to be `*.jpg` -- this doesn't
represent a single file named `*.jpg`, but is a stand-in for all the filenames
ending in `.jpg`. Here, our for loop is able to handle the conversion
comfortably, whether it's just one file or thousands of files.

## Wildcards

The `*.jpg` we have just seen is an example of **wildcard patterns**. Here, `*`
is a **wildcard character** that can match any number of characters, so `*.jpg`
matches `banana.jpg`, `unicorn.jpg`, or even `.jpg` if there happens to be such
a file. The [wildcard expansion](../ref/language.html#wildcard-expansion)
section of the language reference describes wildcards in more details, but `*`
is perhaps what you will use most of the time.

# Multiple values

Something worth remarking with the behavior of `*.jpg` is that it evaluates to
**multiple values**. This means that it becomes multiple elements in a list,
which is what's happening here, but it also becomes multiple arguments when used
in commands. We can see this most clearly with
[the `put` command](../ref/builtin.html#put), which writes each of its argument
as a value output:

```elvish-transcript Terminal - elvish
~> put *.jpg
▶ banana.jpg
▶ unicorn.jpg
```

## Output capture redux

Previously, we have captured the outputs of commands to use as arguments to
other commands, like this:

```elvish-transcript Terminal - elvish
~> * (+ 2 10) 3
▶ (num 36)
```

Here, `(+ 2 10)` outputs a single value, which then gets used as a single
argument.

Some commands in Elvish can output multiple values, and capturing their output
gives us multiple values too. For example, the
[`str:split`](../ref/str.html#str:split) command splits a string around a
separator, outputting one value for each split results:

```elvish-transcript Terminal - elvish
~> str:split , friends,Romands,countrymen
▶ friends
▶ Romands
▶ countrymen
```

We can use these multiple values in the same way we used the multiple values
generated `*.jpg`. For example, we can put them in a list and use that in a for
loop:

```elvish-transcript Terminal - elvish
~> for who [(str:split , friends,Romans,countrymen)] {
     echo 'Hello, '$who'!'
   }
Hello, friends!
Hello, Romans!
Hello, countrymen!
```

Both `+` and `str:split` output values, but what about commands that output
bytes? When we capture their output, each *line* becomes a value. As an example,
<https://dl.elv.sh/INDEX> is a file listing all the files available on the
<https://dl.elv.sh> site. We can use `curl` to request this file and capture the
output:

```elvish-transcript Terminal - elvish
~> for url [(curl -s https://dl.elv.sh/INDEX)] {
     echo 'URL: '$url
   }
URL: https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz
URL: https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz.sha256sum
...
```

For the purpose of examining values, we don't have to put them in a list and use
a for loop. Remember the `put` command, which turns each argument into a value
in its output:

```elvish-transcript Terminal - elvish
~> put (curl -s https://dl.elv.sh/INDEX)
▶ https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz
▶ https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz.sha256sum
...
```

## Lists vs multiple values

A list in Elvish *stores* multiple values, but it's always one value itself. In
some shells and other programming languages, lists can implicitly "become"
multiple values -- that never happens in Elvish.

We have seen how you can turn multiple values into a list simply by wrapping
them inside a pair of `[` and `]`. Conversely, when you have a list and would
like to get all its elements as separate values, you can use the
[`all`](../ref/builtin.html#all) command, which does exactly that:

```elvish-transcript Terminal - elvish
~> all [foo bar]
▶ foo
▶ bar
```

If the list happens to be stored inside a variable `$list`, you can also use the
shorthand `$@list`:

```elvish-transcript Terminal - elvish
~> var list = [foo bar]
~> put $@list
▶ foo
▶ bar
```

# Conclusion

Variables, lists and loops are basic but important
[abstraction](https://en.wikipedia.org/wiki/Abstraction_(computer_science))
mechanisms in programming, and shell scripting is no exception.

In this part, we've learned how to use variables to go beyond simple hardcoded
commands and adapt them to the context they operate in. We've also used loops,
lists and wildcards to repeat operations without even knowing in advance how
many times to repeat them for, and dived into how to make use of multiple
values.

We are now ready for the next part, [Pipelines and IO](pipelines-and-io.html).
