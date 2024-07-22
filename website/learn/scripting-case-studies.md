<!-- toc number-sections -->

This page explains the scripting examples on the [homepage](../), illustrating
the advantages of Elvish scripting, especially when compared to traditional
shells.

For more examples of Elvish features compared to traditional shells, see the
[quick tour](tour.html). For a complete description of the Elvish language, see
the [language reference](../ref/language.html).

# jpg-to-png.elv

This example on the homepage uses
[GraphicsMagick](http://www.graphicsmagick.org) to convert all the `.jpg` files
into `.png` files in the current directory:

```elvish jpg-to-png.elv
for x [*.jpg] {
  gm convert $x (str:trim-suffix $x .jpg).png
}
```

(If you have [ImageMagick](https://imagemagick.org) installed instead, just
replace `gm` with `magick`.)

It's equivalent to the following script in traditional shells:

```sh jpg-to-png.sh
for x in *.jpg
do
  gm convert "$x" "${x%.jpg}".png
done
```

Let's see how the Elvish version compares to the traditional shell version:

-   You don't need to
    [double-quote every variable](https://www.shellcheck.net/wiki/SC2086) to
    prevent unwanted effects. A variable in Elvish always evaluates to one
    value.

-   Instead of `${x%.jpg}`, you write `(str:trim-suffix $x .jpg)`. The latter a
    bit longer, but is easier to remember and understand.

    Moreover, since `str:trim-suffix` is a normal command rather than a special
    operator, it's easy to find its documentation - in the
    [reference doc](../ref/str.html#str:trim-suffix), in the terminal with
    [`doc:show`](../ref/doc.html#doc:show), or by hovering over it in VS Code
    (support for more editors will come).

-   When there is no file that matches `*.jpg`, bash will assign `$x` to the
    pattern `*.jpg` itself, which is most likely not what you want.

    Elvish will throw an exception by default, and you can optionally tell
    Elvish to expand to zero elements with `*[nomatch-ok].jpg`.

-   Perhaps subjectively, Elvish's syntax is more readable: instead of keywords
    like `in`, `do` and `done`, Elvish's [`for`](../ref/language.html#for)
    command doesn't have `in`, and uses familiar punctuation to delimit
    different parts of the command: the list of elements with `[` and `]`, and
    the loop body with `{` and `}`.

This example doesn't go into the advanced capabilities of Elvish, so differences
are small and may seem superficial. However, these small details can quickly add
up, and in general it's much easier to develop and maintain scripts in Elvish.

# update-servers-in-parallel.elv

This example on the homepage shows how you would perform update commands on
multiple servers in parallel:

```elvish update-servers-in-parallel.elv
var hosts = [[&name=a &cmd='apt update']
             [&name=b &cmd='pacman -Syu']]
# peach = "parallel each"
peach {|h| ssh root@$h[name] $h[cmd] } $hosts
```

Let's break down the script:

-   The value of `hosts` is a nested data structure:

    -   The outer `[]` denotes a [list](../ref/language.html#list).

    -   The inner `[]` denotes [maps](../ref/language.html#map) - `&k=v` are
        key-value pairs.

    So the variable `$hosts` contains a list of maps, each map containing the
    key `name` and `cmd`, describing a host.

-   The [`peach`](../ref/builtin.html#peach) command takes:

    -   A [function](../ref/language.html#function), in this case an anonymous
        one. The signature `|h|` denotes that it takes one argument, and the
        body is an `ssh` command using fields from `$h`.

    -   A list, in this case `$hosts`.

As hinted by the comment, `peach` calls the function for every element of the
list in parallel, running these `ssh` commands in parallel. It will also wait
for all the functions to finish before it returns.

In a real-world script, you'll likely want to redirect the output of the `ssh`
command, otherwise the output from the `ssh` commands running in parallel will
interleave each other. This can be done with
[output redirection](../ref/language.html#redirection), not unlike traditional
shells:

```elvish update-servers-in-parallel-v2.elv
var hosts = [[&name=a &cmd='apt update']
             [&name=b &cmd='pacman -Syu']]
peach {|h| ssh root@$h[name] $h[cmd] > ssh-$h[name].log } $hosts
```

With a traditional shell, you can achieve a similar effect with background jobs:

```sh update-servers-in-parallel.sh
ssh root@a 'apt update' > ssh-a.log &
job_a=$!
ssh root@b 'pacman -Syu' > ssh-b.log &
job_b=$!
wait $job_a $job_b
```

However, you will have to manage the lifecycle of the background jobs
explicitly, whereas the
[structured](https://en.wikipedia.org/wiki/Structured_concurrency) nature of
`peach` makes that unnecessary. Alternatively, you can use external commands
such as [GNU Parallel](https://www.gnu.org/software/parallel/) to achieve
parallel execution, but this requires you to learn another tool and structure
your script in a particular way.

Like in most programming languages, data structures in Elvish can be arbitrarily
nested. This allows you to express more complex workflows in a natural way:

```elvish update-servers-in-parallel-v3.elv
var hosts = [[&name=a &cmd='apt update' &dotfiles=[.tmux.conf .gitconfig]]
             [&name=b &cmd='pacman -Syu' &dotfiles=[.vimrc]]]
peach {|h|
  ssh root@$h[name] $h[cmd] > ssh-$h[name].log
  scp ~/(all $h[dotfiles]) root@$h[name]:
} $hosts
```

The expression `(all $h[dotfiles])` evaluates to all the elements of
`$h[dotfiles]` (see documentation for [`all`](../ref/builtin.html#all)), each of
which is then [combined](../ref/language.html#compounding) with `~/`, which
evaluates to the home directory.

Traditional shells tend to have limited support for complex data structures, so
it can get quite tricky to express the same workflow, especially when coupled
with parallel execution.

What's more, you can easily move the definition of `hosts` into a JSON file -
this can be useful if you'd like to share the script with others without
requiring everyone to customize the script:

```json hosts.json
[
  {"name": "a", "cmd": "apt update", "dotfiles": [".tmux.conf", ".gitconfig"]},
  {"name": "b", "cmd": "pacman -Syu", "dotfiles": [".vimrc"]}
]
```

```elvish update-servers-in-parallel-v4.elv
var hosts = (from-json < hosts.json)
peach {|h|
  ssh root@$h[name] $h[cmd] > ssh-$h[name].log
  scp ~/(all $h[dotfiles]) root@$h[name]:
} $hosts
```

Elvish brings the power of data structures and functional programming to your
shell scripting scenarios.

# Catching errors early

The following interaction in the terminal showcases how Elvish is able to catch
errors early:

```elvish-transcript Terminal: elvish
~> var project = ~/project
~> rm -rf $projetc/bin
compilation error: variable $projetc not found
```

The example on the homepage is slightly simplified for brevity. In fact, Elvish
will highlight the exact place of the error, before you even press
<kbd>Enter</kbd> to execute the code:

```ttyshot Terminal: elvish
learn/scripting-case-studies/misspelt-variable
```

In this case, Elvish identifies that the variable name is misspelt and won't
execute the code. Compare this to an interaction in a more traditional shell:

```sh-transcript Terminal: sh
$ project=~/project
$ rm -rf $projetc/bin
[ ...]
```

Traditional shells by default don't treat the misspelt variable as an error and
evaluates it to an empty string instead. As the result, this will start
executing `rm -rf /bin`, possibly with catastrophic consequences.

Elvish's early error checking extends beyond terminal interactions, for example,
suppose you have the following script:

```elvish script-with-error.elv
var project = ~/project
# A function...
fn cleanup-bin {
  rm $projetc/bin
}
# More code...
```

If you try to run this script with `elvish script-with-error.elv`, Elvish will
find the misspelt variable name within the function `cleanup-bin`, and refuses
to execute any code from the script.

Elvish's early error checking can help you prevent a lot of bugs from simple
typos. There are more places Elvish checks for errors, and more checks are being
added.

# Command failures

The following terminal interaction shows Elvish's behavior when a command fails:

```elvish-transcript Terminal: elvish
~> gm convert a.jpg a.png; rm a.jpg
gm convert: Failed to convert a.jpg
Exception: gm exited with 1
  [tty 1]:1:1-22: gm convert a.jpg a.png; rm a.jpg
# "rm a.jpg" is NOT executed
```

Like traditional shells, you can connect multiple commands together with either
`;` (as in this example) or newlines, and Elvish will run them one after
another.

However, unlike traditional shells, if any command fails with a non-zero exit
code, Elvish defaults to aborting execution. As the output indicates, this is
part of a general mechanism of exceptions, which can be
[caught](../ref/language.html#try) and
[inspected](../ref/language.html#exception). (If you're familiar with the
`set -e` mechanism in traditional shells, Elvish's behavior is similar, but
without its [many](https://david.rothlis.net/shell-set-e/)
[flaws](http://mywiki.wooledge.org/BashFAQ/105).)

This early abortion behavior is a much safer default for scripting. In
particular, it's almost certainly what you want for CI/CD scripts. For example,
consider the following script:

```elvish
./run-tests
./run-linters
```

If `./run-tests` fails, Elvish will abort the entire script, but a traditional
shell will happily proceed and only fail if `./run-linters` fails.
