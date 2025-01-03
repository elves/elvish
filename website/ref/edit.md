<!-- toc -->

@module edit

The `edit:` module is the interface to the Elvish editor.

Function usages are given in the same format as in the
[reference for the builtin module](builtin.html).

*This document is incomplete.*

# Overview

## Modes and Submodules

The Elvish editor has different **modes**, and exactly one mode is active at the
same time. Each mode has its own UI and keybindings. For instance, the default
**insert mode** lets you modify the current command. The **completion mode**
(triggered by <kbd>Tab</kbd> by default) shows you all candidates for
completion, and you can use arrow keys to navigate those candidates.

```ttyshot
ref/edit/completion-mode
```

Each mode has its own submodule under `edit:`. For instance, builtin functions
and configuration variables for the completion mode can be found in the
`edit:completion:` module.

The primary modes supported now are `insert`, `command`, `completion`,
`navigation`, `history`, `histlist`, `location`, and `lastcmd`. The last 4 are
"listing modes", and their particularity is documented below.

## Prompts

Elvish has two prompts: the (normal) left-hand prompt and the right-side prompt
(rprompt). Most of this section only documents the left-hand prompt, but API for
rprompt is the same other than the variable name: just replace `prompt` with
`rprompt`.

To customize the prompt, assign a function to `edit:prompt`. The function may
write value outputs or byte outputs:

-   Value outputs may be either strings or `styled` values; they are joiend with
    no spaces in between.

-   Byte outputs are output as-is, including any newlines. Any
    [SGR escape sequences](https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters)
    included in the byte outputs will be parsed, but any other escape sequences
    or control character will be removed.

If you mix value and byte outputs, the order in which they appear is
non-deterministic.

Prefer using `styled` to output styled text; the support for SGR escape
sequences is mostly for compatibility with external cross-shell prompts.

The default prompt and rprompt are equivalent to:

```elvish
set edit:prompt = { tilde-abbr $pwd; put '> ' }
set edit:rprompt = (constantly (styled (whoami)@(hostname) inverse))
```

Some more examples:

```elvish-transcript
~> set edit:prompt = { tilde-abbr $pwd; styled '> ' green }
~> # ">" is now green
~> set edit:prompt = { echo '$' }
$
# Cursor will be on the next line as `echo` outputs a trailing newline
```

### Stale Prompt

Elvish never waits for the prompt function to finish. Instead, the prompt
function is always executed on a separate thread, and Elvish updates the screen
when the function finishes.

However, this can be misleading when the function is slow: this means that the
prompt on the screen may not contain the latest information. To deal with this,
if the prompt function does not finish within a certain threshold - by default
0.2 seconds, Elvish marks the prompt as **stale**: it still shows the old stale
prompt content, but transforms it using a **stale transformer**. The default
stale transformer applies reverse-video to the whole prompt.

The threshold is customizable with `$edit:prompt-stale-threshold`; it specifies
the threshold in seconds.

The transformer is customizable with `$edit:prompt-stale-transform`. It is a
function; the function is called with one argument, a `styled` text, and the
output is interpreted in the same way as prompt functions. Some examples are:

```elvish
# The following effectively disables marking of stale prompt.
set edit:prompt-stale-transform = {|x| put $x }
# Show stale prompts in inverse; equivalent to the default.
set edit:prompt-stale-transform = {|x| styled $x inverse }
# Gray out stale prompts.
set edit:prompt-stale-transform = {|x| styled $x bright-black }
```

To see the transformer in action, try the following example (assuming default
`$edit:prompt-stale-transform`):

```elvish
var n = 0
set edit:prompt = { sleep 2; put $n; set n = (+ $n 1); put ': ' }
set edit:-prompt-eagerness = 10 # update prompt on each keystroke
set edit:prompt-stale-threshold = 0.5
```

And then start typing. Type one character; the prompt becomes inverse after 0.5
second: this is when Elvish starts to consider the prompt as stale. The prompt
will return normal after 2 seconds, and the counter in the prompt is updated:
this is when the prompt function finishes.

Another thing you will notice is that, if you type a few characters quickly (in
less than 2 seconds, to be precise), the prompt is only updated twice. This is
because Elvish never does two prompt updates in parallel: prompt updates are
serialized. If a prompt update is required when the prompt function is still
running, Elvish simply queues another update. If an update is already queued,
Elvish does not queue another update. The reason why exactly two updates happen
in this case, and how this algorithm ensures freshness of the prompt is left as
an exercise to the reader.

### Prompt Eagerness

The occasions when the prompt should get updated can be controlled with
`$edit:-prompt-eagerness`:

-   The prompt is always updated when the editor becomes active -- when Elvish
    starts, or a command finishes execution, or when the user presses Enter.

-   If `$edit:-prompt-eagerness` >= 5, it is updated when the working directory
    changes.

-   If `$edit:-prompt-eagerness` >= 10, it is updated on each keystroke.

The default value is 5.

### RPrompt Persistency

By default, the rprompt is only shown while the editor is active: as soon as you
press Enter, it is erased. If you want to keep it, simply set
`$edit:rprompt-persistent` to `$true`:

```elvish
set edit:rprompt-persistent = $true
```

## Keybindings

Each mode has its own keybinding, accessible as the `binding` variable in its
module. For instance, the binding table for insert mode is
`$edit:insert:binding`. To see current bindings, simply print the binding table:
`pprint $edit:insert:binding` (replace `insert` with any other mode).

The global key binding table, `$edit:global-binding` is consulted when a key is
not handled by the active mode.

A binding tables is simply a map that maps keys to functions. For instance, to
bind `Alt-x` in insert mode to exit Elvish, simply do:

```elvish
set edit:insert:binding[Alt-x] = { exit }
```

Outputs from a bound function always appear above the Elvish prompt. You can see
this by doing the following:

```elvish
set edit:insert:binding[Alt-x] = { echo 'output from a bound function!' }
```

and press <kbd>Alt-x</kbd> in insert mode. It allows you to put debugging
outputs in bound functions without messing up the terminal.

Internally, this is implemented by connecting their output to a pipe. This does
the correct thing in most cases, but if you are sure you want to do something to
the terminal, redirect the output to `/dev/tty`. Since this will break Elvish's
internal tracking of the terminal state, you should also do a full redraw with
`edit:redraw &full=$true`. For instance, the following binds <kbd>Ctrl-L</kbd>
to clearing the terminal:

```elvish
set edit:insert:binding[Ctrl-L] = { clear > /dev/tty; edit:redraw &full=$true }
```

(The same functionality is already available as a builtin, `edit:clear`.)

Bound functions have their inputs redirected to /dev/null.

### Format of Keys

A key consists of a **key name**, preceded by zero or more **key modifiers**.
Both are case-sensitive.

The key name may be either:

-   A simple character, such as `x`.

    Most of the time you should use the lower-case form of letters, except for
    Ctrl- bindings, which usually require the upper-case form (so `Alt-x` but
    `Ctrl-X`). This quirk may go away in future.

-   A function key from these symbols:

    ```
    F1  F2  F3  F4  F5  F6  F7  F8  F9  F10  F11  F12
    Up  Down  Right  Left
    Home  Insert  Delete  End  PageUp  PageDown
    Tab  Enter  Backspace
    ```

    Among these, `Tab` and `Enter` are equivalent to their ASCII counterparts
    `"\t"` and `"\n"`.

Key modifiers are one of the following:

```
A  Alt
C  Ctrl
M  Meta
S  Shift
```

Modifiers end with either a `-` or `+`, and can be stacked. Examples:

-   `S-F1`

-   `Ctrl-X` or

-   `Alt+Enter`

-   `C+A-X`.

**Note:** The `Shift` modifier is only applicable to function keys such as `F1`.
You cannot write `Shift-m` as a synonym for `M`.

You may not actually be able to use the full range of possible keys for several
reasons:

-   Some `Ctrl-[letter]` keys have special functions, like `Ctrl-C`, `Ctrl-Z`,
    `Ctrl-S` and `Ctrl-Z`.

-   Some `Ctrl-[letter]` keys are equivalent to single keypresses, like `Ctrl-J`
    (equivalent to `Enter`) and `Ctrl-I` (equivalent to `Tab`).

-   Keys involving multiple modifiers may not be supported by the terminal
    emulator, especially when the base key is a function key.

### Listing Modes

The modes `histlist`, `location` and `lastcmd` are all **listing modes**: They
all show a list, and you can filter items and accept items.

Because they are very similar, you may want to change their bindings at the same
time. This is made possible by the `$edit:listing:binding` binding table
(`listing` is not a "real" mode but an "abstract" mode). These modes still have
their own binding tables like `$edit:histlist:binding`, and bindings there have
higher precedence over those in the shared `$edit:listing:binding` table.

Moreover, there are a lot of builtin functions in the `edit:listing` module like
`edit:listing:down` (for moving down selection). They always apply to whichever
listing mode is active.

### Caveat: Bindings to Start Modes

Note that keybindings to **start** modes live in the binding table of the insert
mode, not the target mode. For instance, if you want to be able to use
<kbd>Alt-l</kbd> to start location mode, you should modify
`$edit:insert:binding[Alt-l]`:

```elvish
set edit:insert:binding[Alt-l] = { edit:location:start }
```

One tricky case is the history mode. You can press <kbd>▲︎</kbd> to start
searching for history, and continue pressing it to search further. However, when
the first press happens, the editor is in insert mode, while with subsequent
presses, the editor is in history mode. Hence this binding actually relies on
two entries, `$edit:insert:binding[Up]` and `$edit:history:binding[Up]`.

So for instance if you want to be able to use <kbd>Ctrl-P</kbd> for this, you
need to modify both bindings:

```elvish
set edit:insert:binding[Ctrl-P] =  { edit:history:start }
set edit:history:binding[Ctrl-P] = { edit:history:up }
```

## Filter DSL

The completion, history listing, location and navigation modes all support
filtering the items to show using a filter DSL. It uses a small subset of
Elvish's expression syntax, and can be any of the following:

-   A literal string (barewords and single-quoted or double-quoted strings all
    work) matches items containing the string. If the string is all lower case,
    the match is done case-insensitively; otherwise the match is case-sensitive.

-   A list `[re $string]` matches items matching the regular expression
    `$string`. The `$string` must be a literal string.

-   A list `[and $expr...]` matches items matching all of the `$expr`s.

-   A list `[or $expr...]` matches items matching any of the `$expr`s.

If the filter contains multiple expressions, they are ANDed, as if surrounded by
an implicit `[and ...]`.

## Completion API

### Argument Completer

There are two types of completions in Elvish: completion for internal data and
completion for command arguments. The former includes completion for variable
names (e.g. `echo $`<kbd>Tab</kbd>) and indices (e.g.
`echo $edit:insert:binding[`<kbd>Tab</kbd>). These are the completions that
Elvish can provide itself because they only depend on the internal state of
Elvish.

The latter, in turn, is what happens when you type e.g. `cat`<kbd>Tab</kbd>.
Elvish cannot provide completions for them without full knowledge of the
command.

Command argument completions are programmable via the
`$edit:completion:arg-completer` variable. When Elvish is completing an argument
of command `$x`, it will call the value stored in
`$edit:completion:arg-completer[$x]`, with all the existing arguments, plus the
command name in the front.

For example, if the user types `man 1`<kbd>Tab</kbd>, Elvish will call:

```elvish
$edit:completion:arg-completer[man] man 1
```

If the user is starting a new argument when hitting <kbd>Tab</kbd>, Elvish will
call the completer with a trailing empty string. For instance, if you do
`man 1`<kbd>Space</kbd><kbd>Tab</kbd>, Elvish will call:

```elvish
$edit:completion:arg-completer[man] man 1 ""
```

The output of this call becomes candidates. There are several ways of outputting
candidates:

-   Writing byte output, e.g. "echo cand1; echo cand2". Each line becomes a
    candidate. This has the drawback that you cannot put newlines in candidates.
    Only use this if you are sure that you candidates will not contain newlines
    -- e.g. package names, usernames, but **not** file names, etc..

-   Write strings to value output, e.g. "put cand1 cand2". Each string output
    becomes a candidate.

-   Use the `edit:complex-candidate` command, e.g.:

    ```elvish
    edit:complex-candidate &code-suffix='' &display=$stem' ('$description')'  $stem
    ```

    See [`edit:complex-candidate`]() for the full description of the arguments
    is accepts.

After receiving your candidates, Elvish will match your candidates against what
the user has typed. Hence, normally you don't need to (and shouldn't) do any
matching yourself.

That means that in many cases you can (and should) simply ignore the last
argument to your completer. However, they can be useful for deciding what
**kind** of things to complete. For instance, if you are to write a completer
for `ls`, you want to see whether the last argument starts with `-` or not: if
it does, complete an option; and if not, complete a filename.

Here is a very basic example of configuring a completer for the `apt` command.
It only supports completing the `install` and `remove` command and package names
after that:

```elvish
use re

var all-packages = [(apt-cache search '' | re:awk {|0 1 @rest| put $1 })]

set edit:completion:arg-completer[apt] = {|@args|
    var n = (count $args)
    if (== $n 2) {
        # apt x<Tab> -- complete a subcommand name
        put install uninstall
    } elif (== $n 3) {
        put $@all-packages
    }
}
```

Here is another slightly more complex example for the `git` command. It supports
completing some common subcommands and then branch names after that:

```elvish
use re

fn all-git-branches {
    # Note: this assumes a recent version of git that supports the format
    # string used.
    git branch -a --format="%(refname:strip=2)" | re:awk {|0 1 @rest| put $1 }
}

var common-git-commands = [
  add branch checkout clone commit diff init log merge
  pull push rebase reset revert show stash status
]

set edit:completion:arg-completer[git] = {|@args|
    var n = (count $args)
    if (== $n 2) {
        put $@common-git-commands
    } elif (>= $n 3) {
        all-git-branches
    }
}
```

### Matcher

As stated above, after the completer outputs candidates, Elvish matches them
with them with what the user has typed. For clarity, the part of the user input
that is relevant to tab completion is called for the **seed** of the completion.
For instance, in `echo x`<kbd>Tab</kbd>, the seed is `x`.

Elvish first indexes the matcher table -- `$edit:completion:matcher` -- with the
completion type to find a **matcher**. The **completion type** is currently one
of `variable`, `index`, `command`, `redir` or `argument`. If the
`$edit:completion:matcher` lacks the suitable key,
`$edit:completion:matcher['']` is used.

Elvish then calls the matcher with one argument -- the seed, and feeds the
*text* of all candidates to the input. The mather must output an identical
number of booleans, indicating whether the candidate should be kept.

As an example, the following code configures a prefix matcher for all completion
types:

```elvish
set edit:completion:matcher[''] = {|seed| each {|cand| has-prefix $cand $seed } }
```

Elvish provides three builtin matchers, `edit:match-prefix`, `edit:match-substr`
and `edit:match-subseq`. In addition to conforming to the matcher protocol, they
accept two options `&ignore-case` and `&smart-case`. For example, if you want
completion of arguments to use prefix matching and ignore case, use:

```elvish
set edit:completion:matcher[argument] = {|seed| edit:match-prefix $seed &ignore-case=$true }
```

The default value of `$edit:completion:matcher` is `[&''=$edit:match-prefix~]`,
hence that candidates for all completion types are matched by prefix.

## Hooks

Hooks are functions that are executed at certain points in time. In Elvish this
functionality is provided by variables that are a list of functions.

**NOTE**: Hook variables may be initialized with a non-empty list, and you may
have modules that add their own hooks. In general you should append to a hook
variable rather than assign a list of functions to it. That is, rather than
doing `set edit:some-hook = [ { put 'I ran' } ]` you should do
`set edit:some-hook = [ $@hook-var { put 'I ran' } ]`.

These are the editor/REPL hooks:

-   [`$edit:before-readline`](https://elv.sh/ref/edit.html#editbefore-readline):
    The functions are called before the editor runs. Each function is called
    with no arguments.

-   [`$edit:after-readline`](https://elv.sh/ref/edit.html#editafter-readline):
    The functions are called after the editor accepts a command for execution.
    Each function is called with a sole argument: the line just read.

-   [`$edit:after-command`](https://elv.sh/ref/edit.html#editafter-command): The
    functions are called after the shell executes the command you entered
    (typically by pressing the `Enter` key). Each function is called with a sole
    argument: a map that provides information about the executed command. This
    hook is also called after your interactive RC file is executed and before
    the first prompt is output.

Example usage:

```elvish
set edit:before-readline = [{ echo 'going to read' }]
set edit:after-readline = [{|line| echo 'just read '$line }]
set edit:after-command = [{|m| echo 'command took '$m[duration]' seconds' }]
```

Given the above hooks...

1.  Every time you accept a chunk of code (normally by pressing Enter)
    `just read` is printed.

2.  At the very beginning of an Elvish session, or after a chunk of code is
    handled, `going to read` is printed.

3.  After each non empty chunk of code is accepted and executed the string
    "command took ... seconds\` is output.

## Word types

The editor supports operating on entire "words". As intuitive as the concept of
"word" is, there is actually no single definition for the concept. The editor
supports the following three definitions of words:

-   A **big word**, or simply **word**, is a sequence of non-whitespace
    characters. This definition corresponds to the concept of "WORD" in vi.

-   A **small word** is a sequence of alphanumerical characters ("alnum small
    word"), or a sequence of non-alphanumerical, non-whitespace characters
    ("punctuation small word"). This definition corresponds to the concept of
    "word" in vi and zsh.

-   An **alphanumerical word** is a sequence of alphanumerical characters. This
    definition corresponds to the concept of "word" in bash.

Whitespace characters are those with the Unicode
[Whitespace](https://en.wikipedia.org/wiki/Whitespace_character#Unicode)
property. Alphanumerical characters are those in the Unicode Letter or Number
category.

A **word boundary** is an imaginary zero-length boundary around a word.

To see the difference between these definitions, consider the following string:
`abc++ /* xyz`:

-   It contains three (big) words: `abc++`, `/*` and `xyz`.

-   It contains four small words, `abc`, `++`, `/*` and `xyz`. Among them, `abc`
    and `xyz` are alnum small words, while `++` and `/*` are punctuation small
    words.

-   It contains two alnum words, `abc` and `xyz`.

## Autofix

The editor can identify **autofix** commands to fix some errors in the code.

For example, if you try to use a command from [the `str:` module](str.html)
without importing it, the editor will offer `use str` as an autofix command:

```ttyshot
ref/edit/autofix
```

As seen above, autofixes are also applied automatically by
[`edit:completion:smart-start`]() (the default binding for <kbd>Tab</kbd>) and
[`edit:smart-enter`]() (the default binding for <kbd>Enter</kbd>).
