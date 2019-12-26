<!-- toc -->

If you come from other shells, hopefully the following recipes will get you
started quickly:

# UI Recipes

-   Put your startup script in `~/.elvish/rc.elv`. There is no `alias` yet, but
    you can achieve the goal by defining a function:

    ```elvish
    fn ls [@a]{ e:ls --color $@a }
    ```

    The `e:` prefix (for "external") ensures that the external command named
    `ls` will be called. Otherwise this definition will result in infinite
    recursion.

-   The left and right prompts can be customized by assigning functions to
    `edit:prompt` and `edit:rprompt`. Their outputs are concatenated (with no
    spaces in between) before being used as the respective prompts. The
    following simulates the default prompts but uses fancy Unicode:

    ```elvish
    # "tilde-abbr" abbreviates home directory to a tilde.
    edit:prompt = { tilde-abbr $pwd; put '❱ ' }
    # "constantly" returns a function that always writes the same value(s) to
    # output; "edit:styled" writes styled output.
    edit:rprompt = (constantly (edit:styled (whoami)✸(hostname) inverse))
    ```

    Here is a terminalshot of the alternative prompts:

    $ttyshot unicode-prompts

-   Press <span class="key">▲&#xfe0e;</span> to search through history. It uses
    what you have typed to do prefix match. To cancel, press <span
    class="key">Escape</span>.

    $ttyshot history-mode

-   Press <span class="key">Tab</span> to start completion. Use arrow keys
    <span class="key">▲&#xfe0e;</span> <span class="key">▼&#xfe0e;</span>
    <span class="key">◀&#xfe0e;</span> <span class="key">▶&#xfe0e;</span> or
    <span class="key">Tab</span> and <span class="key">Shift-Tab</span> to
    select the candidate. Press <span class="key">Enter</span>, or just continue
    typing to accept. To cancel, press <span
    class="key">Escape.</span> It even comes with a scrollbar! :) In fact, all
    interactive modes show a scrollbar when there is more output to see.

    $ttyshot completion-mode

-   You can make completion case-insensitive with the following code:

    ```elvish
    edit:-matcher[''] = [p]{ edit:match-prefix &ignore-case $p }
    ```

    You can also make the completion use "smart case" by changing `&ignore-case`
    to `&smart-case`. This means that if your pattern is entirely lower case it
    ignores case, otherwise it's case sensitive.

-   <a name="navigation-mode"></a>Press <span class="key">Ctrl-N</span> to start
    the builtin filesystem navigator, aptly named "navigation mode." Use arrow
    keys to navigate. <span class="key">Enter</span> inserts the selected
    filename to your command line. If you want to insert the filename and stay
    in the mode (e.g. when you want to insert several filenames), use <span
    class="key">Alt-Enter</span>.

    You can continue typing your command when you are in navigation mode. Press
    <span class="key">Ctrl-H</span> to toggle hidden files; and like in other
    modes, <span class="key">Escape</span> gets you back to the default (insert)
    mode.

    $ttyshot navigation-mode

-   Try typing `echo [` and press <span class="key">Enter</span>. Elvish knows
    that the command is unfinished due to the unclosed `[` and inserts a newline
    instead of accepting the command. Moreover, common errors like syntax errors
    and missing variables are highlighted in real time.

-   Elvish remembers which directories you have visited. Press <span
    class="key">Ctrl-L</span> to list visited directories. Like in completion,
    use arrow keys <span class="key">▲&#xfe0e;</span>
    <span class="key">▼&#xfe0e;</span> or <span class="key">Tab</span> and
    <span class="key">Shift-Tab</span> to select a directory and use Enter to
    `cd` into it. Press <span
    class="key">Escape</span> to cancel.

    $ttyshot location-mode

    Type to filter:

    $ttyshot location-mode-filter

    The filtering algorithm is tailored for matching paths; you need only type a
    prefix of each component. In the screenshot, x/p/v matches
    **x**iaq/**p**ersistent/**v**ector.

-   Elvish doesn't support history expansion like `!!`. Instead, it has a "last
    command mode" offering the same functionality, triggered by <span
    class="key">Alt-1</span> by default (resembling how you type `!` using
    <span class="key">Shift-1</span>). In this mode, you can pick individual
    arguments from the last command using numbers, or the entire command by
    typing <span class="key">Alt-1</span> again.

    This is showing me trying to fix a forgotten `sudo`:

    $ttyshot lastcmd

# Language Recipes

-   Lists look like `[a b c]`, and maps look like `[&key1=value1 &key2=value2]`.
    Unlike other shells, a list never expands to multiple words, unless you
    explicitly explode it by prefixing the variable name with `@`:

    ```elvish-transcript
    ~> li = [1 2 3]
    ~> put $li
    ▶ [1 2 3]
    ~> put $@li
    ▶ 1
    ▶ 2
    ▶ 3
    ~> map = [&k1=v1 &k2=v2]
    ~> echo $map[k1]
    v1
    ```

-   Environment variables live in a separate `E:` (for "environment") namespace
    and must be explicitly qualified:

    ```elvish-transcript
    ~> put $E:HOME
    ▶ /home/xiaq
    ~> E:PATH = $E:PATH":/bin"
    ```

-   You can manipulate search paths through the special list `$paths`, which is
    synced with `$E:PATH`:

    ```elvish-transcript
    ~> echo $paths
    [/bin /sbin]
    ~> paths = [/opt/bin $@paths /usr/bin]
    ~> echo $paths
    [/opt/bin /bin /sbin /usr/bin]
    ~> echo $E:PATH
    /opt/bin:/bin:/sbin:/usr/bin
    ```

-   You can manipulate the keybinding in the default insert mode through the map
    `$edit:insert:binding`. For example, this binds
    <span class="key">Ctrl-L</span> to clearing the terminal:

    ```elvish
    edit:insert:binding[Ctrl-L] = { clear > /dev/tty }
    ```

    Use `pprint $edit:insert:binding` to get a nice (albeit long) view of the
    current keybinding.

    **NOTE**: Bindings for letters modified by Alt are case-sensitive. For
    instance, `Alt-a` means pressing `Alt` and `A`, while `Alt-A` means pressing
    `Alt`, `Shift` and `A`. This will probably change in the future.

-   There is no interpolation inside double quotes (yet). For example, the
    output of `echo "$user"` is simply the string `$user`. Use implicit string
    concatenation to build strings:

    ```elvish-transcript
    ~> name = xiaq
    ~> echo "My name is "$name"."
    My name is xiaq.
    ```

    Sometimes string concatenation will force you to use string literals instead
    of barewords:

    ```elvish-transcript
    ~> noun = language
    ~> echo $noun's'
    languages
    ```

    You cannot write `s` as a bareword because Elvish would think you are trying
    to write the variable `$nouns`. It's hard to make such mistakes when working
    interactively, as Elvish highlights variables and complains about
    nonexistent variables as you type.

-   Double quotes support C-like escape sequences (`\n` for newline, etc.):

    ```elvish-transcript
    ~> echo "a\nb"
    a
    b
    ```

    **NOTE**: If you run `echo "a\nb"` in bash or zsh, you might get the same
    result (depending on the value of some options), and this might lead you to
    believe that they support C-like escape sequences in double quotes as well.
    This is not the case; bash and zsh preserve the backslash in double quotes,
    and it is the `echo` builtin command that interpret the escape sequences.
    This difference becomes apparent if you change `echo` to `touch`: In Elvish,
    `touch "a\nb"` creates a file whose name has a newline; while in bash or
    zsh, it creates a file whose name contains a backslash followed by `n`.

-   Elementary floating-point arithmetic as well as comparisons are builtin,
    with a prefix syntax:

    ```elvish-transcript
    ~> + 1 2
    ▶ 3
    ~> / (* 2 3) 4
    ▶ 1.5
    ~> > 1 2
    ▶ $false
    ~> < 1 2
    ▶ $true
    ```

    **NOTE**: Elvish has special parsing rules to recognize `<` and `>` as
    command names. That means that you cannot put redirections in the beginning;
    bash and zsh allows `< input cat`, which is equivalent to `cat < input`; in
    Elvish, you can only use the latter syntax.

-   Functions are defined with `fn`. You can name arguments:

    ```elvish-transcript
    ~> fn square [x]{
         * $x $x
         }
    ~> square 4
    ▶ 16
    ```

-   Output of some builtin commands start with a funny `▶`. It is not part of
    the output itself, but shows that such commands output a stream of values
    instead of bytes. As such, their internal structures as well as boundaries
    between values are preserved. This allows us to manipulate structured data
    in the shell.

    Read [unique semantics](unique-semantics.html) for details.

-   When calling Elvish commands, options use the special `&option=value`
    syntax. For instance, the `echo` builtin command supports a `sep` option for
    specifying an alternative separator:

    ```elvish-transcript
    ~> echo &sep="," a b c
    a,b,c
    ```

    The mixture of options and arguments is a classical problem in traditional
    shell programming. For instance, if you want to print a file whose name is
    in `$x`, `cat $x` is the obvious thing to do, but it does not do this
    reliably -- if `$x` starts with `-` (e.g. `-v`), `cat` thinks that it is an
    option. The correct way is `cat -- $x`.

    Elvish commands are free from this problem. However, the option facility is
    only available to builtin commands and user-defined functions, not external
    commands, meaning that you still need to do `cat -- $x`.

    In principle, it is possible to write safe wrapper for external commands and
    there is a [plan](https://github.com/elves/elvish/issues/371) for this.
