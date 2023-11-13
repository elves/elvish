<!-- toc -->

# Customizing Interactive Elvish

Elvish has a large number of features that make it a friendly interactive shell
without any customization. You may, however, find features missing that you are
used to or are otherwise useful to most Elvish users. This is sometimes because
the Elvish builtin commands, as a rule, do not special-case their string
arguments. The core Elvish implementation is meant to resemble a traditional
programming language. This means that functionality you might be used to, such
as `cd -` to change to the previous CWD, doesn't exist. Many usability features
are instead provided by things like [location mode](tour.html#directory-history)
(aka "directory history") which, while technically implemented by a command you
can type, are meant to be invoked via a key binding to that command.

All of the following suggestions involve modifying your
[Elvish interactive configuration](../ref/command.html#rc-file). You may find
that adding some of the suggestions below to your Elvish config results in a
better user experience but none of them are necessary to use Elvish. You are
free to use none, some, or all of the customizations below.

Unless otherwise noted each customization should work on Windows and all Unix
systems. However, in some cases you may need to install additional software;
e.g., the `fzf` program. Also, all of the customizations assume you have
imported the required modules in your *rc.elv*; e.g., `use os`.

There are also many
[third-party modules](https://github.com/elves/awesome-elvish) you may find
useful.

# Import all builtin modules

You will eventually find yourself wanting to interactively use one of the
builtin commands such as [`str:join`](../ref/str.html#str:join). This is easier
if you have already imported those modules. I put the following near the top of
my Elvish config so all of those useful builtin commands are readily available.

```elvish
# Standard elvish modules we always want to be readily available.
use builtin
use doc
# use edit is implicit and cannot be explicitly imported
use epm
use file
use flag
use math
use os
use path
use platform
use re
use runtime
use store
use str
# Conditionally import the `unix` module into the interactive namespace so
# that we don't get an exception trying to import it on Windows.
if $platform:is-unix {
  use unix
  edit:add-var unix: $unix:
}
```

Many users will also want to include
[`use readline-binding`](../ref/readline-binding.html) for its side-effect of
installing Emacs style key bindings similar to those provided by popular shells
like Bash. However, if you do so I recommend also adding the following since
location and navigation mode are far more useful than the clear screen and next
newest command history behavior. This is obviously highly personal so feel free
to experiment with what works best for you and do read the
[`readline-binding`](https://src.elv.sh/pkg/mods/readlinebinding/readline-binding.elv)
implementation.

```
# Reinstate the non-readline default binding for these behaviors.
set edit:insert:binding[Ctrl-L] = $edit:location:start~
set edit:insert:binding[Ctrl-N] = $edit:navigation:start~
```

# How to get `cd -` behavior

Traditional POSIX shells, like Bash, special-case `cd -` to change to the
previously visited directory rather than change to a literal `-` directory.
Elvish does not implement that special-case. This is because the Elvish builtin
commands, as a rule, do not special-case their string arguments.

Do not despair. Adding the following to your Elvish config means you can press
Ctrl-L then Enter to switch to the previous directory (saving three keystrokes).
Also, you can press Ctrl-L and quickly select one of the other five most
recently visited directories. Obviously you can change the stack size (the
constant `4` in the following code) to increase or decrease the size of the
recently visited directory stack. See the description of
[location mode](tour.html#directory-history) (aka, directory history).

```
# Support the equivalent of `cd -` in location mode (Ctrl-L).
set before-chdir = (conj $before-chdir ^
  {|_| var prev = [(take 4 [(
            each {|dir| if (not (eq $pwd $dir)) { put $dir } } ^
                 $edit:location:pinned )] )]
    set edit:location:pinned = [$pwd $@prev]
  })
```

P.S.: Note that `cd -` in Bash is not just an interactive shell shortcut. It
also works in non-interactive Bash scripts. Which is a source of potential bugs
since any string passed to the `cd` builtin might produce unexpected behavior if
it is a magic string unless "escaped". This is one reason why Elvish does not
implement that behavior in its `cd` builtin. Elvish prefers to be safe from
unexpected behavior by not recognizing magic strings.

# Using an external editor to edit commands

If you are working on a complicated command, especially if it spans multiple
lines, you might find yourself frustrated by the Elvish command editor. Adding
the following statements to your Elvish config makes it easy to use your
preferred editor to modify the command by pressing Alt-e or Alt-v or whatever
key sequence you prefer. I use Alt-e and Alt-v for the mnemonic to "emacs" (or
editor) and "vi". If you exit your editor with a zero (success) status your
modified text will replace the current command. If you exit your editor with a
non-zero status then the current command will be unchanged.

```elvish
fn external-edit-command {
  var temp-file = (os:temp-file '*.elv')
  echo $edit:current-command > $temp-file
  try {
    # This assumes $E:EDITOR is an absolute path. If you prefer to use
    # just the bare command and have it resolved when this is run use
    # `(external $E:EDITOR)` or something like `e:vim`.
    $E:EDITOR $temp-file[name] </dev/tty >/dev/tty 2>&1
    set edit:current-command = (str:trim-right (slurp < $temp-file[name]) " \n")
  } finally {
    file:close $temp-file
    os:remove $temp-file[name]
  }
}

# Arrange for Alt-e and Alt-v to edit the current command buffer using my
# preferred external editor.
set edit:insert:binding[Alt-e] = $external-edit-command~
set edit:insert:binding[Alt-v] = $external-edit-command~
```

# Add a `help` command

Elvish has a [`doc`](../ref/doc.html) module for searching and showing
documentation of its builtin commands. However, the `doc` commands are low level
and somewhat awkward for interactive use. The following block of code defines a
`help` command that provides a friendlier mechanism for searching and displaying
documentation of the builtin commands. Try something like `help string echo`,
`help math:`, `help platform:os`, or `help &search builtin:` (the last one finds
all uses of `builtin:` in the documentation rather than display the commands in
that module).

```elvish
fn help {|&search=$false @args|
  if (and (eq $search $false) (== 1 (count $args))) {
    try {
      doc:show $args[0]
    } catch {
      try {
        doc:show '$'$args[0]
      } catch {
        doc:find $args[0]
      }
    }
  } else {
    doc:find $@args
  }
}
```

# Define a clear screen key

Traditional POSIX shells define Ctrl-L to clear the screen. Elvish provides a
command to do so, [`edit:clear`](../ref/edit.html#edit:clear), but it isn't
mapped to a key by default unless you include
[`use readline-binding`](../ref/readline-binding.html) in your config.

```elvish
# Provide a "clear screen" key. I'll rarely use this but define it in case I
# find myself wanting it. Note that I don't use Ctrl-L because that is mapped
# to edit:location:start by default and I use location mode far more often
# than I want to clear the screen.
set edit:insert:binding[Alt-l] = $edit:clear~
```

Note that if you press Alt-l when typing a command it will clear the screen and
show you a new prompt with your unfinished command. Obviously you can customize
the key binding above to Ctrl-l, or some other key, if that is your preference.

# Sync global command history with local history

Elvish has a global persistent command history. Every time you execute a command
interactively the equivalent of [store:add-cmd](../ref/store.html#store:add-cmd)
is executed. That global command history is normally only read when an
interactive Elvish instance begins running. After that any commands run in other
Elvish instances are not visible unless you start a new Elvish instance. You can
get visibility to those commands by running
[`edit:history:fast-forward`](../ref/edit.html#edit:history:fast-forward). That,
however, is a lot to type and hard to remember. So I define the following
function to give the long command a significantly shorter name. You could also
do this using a [command abbreviation](../ref/edit.html#$edit:command-abbr).
Obviously you can name this shortcut anything you wish and bind it to a key.

```elvish
# Sync our local history with the global history. Equivalent to `history
# read` or `history -r` in other shells; hence the `hr` shortcut name.
fn hr {||
    edit:history:fast-forward
}
```

# Use `fzf` to browse command history

If you use the [`fzf`](https://github.com/junegunn/fzf) program (a command-line
fuzzy finder) to select files or items from a list you might prefer to use it
for selecting commands from your history rather than the default
[`edit:histlist:start`](../ref/edit.html#edit:histlist:start) command. The
following function shows how to do so efficiently. This is fast enough that you
are unlikely to notice any slowness compared to the builtin command history
listing mode.

Note that I prefer the behavior provided by the `--exact` option. If you want
fuzzy searching remove that option from the `fzf` command. Similarly, you may
prefer other `fzf` options; e.g., a dark color mode. You are free to tweak the
`fzf` invocation to suit your preferences.

```elvish
# Filter the command history through the fzf program. This is normally bound
# to Ctrl-R.
fn history {||
  var new-cmd = (
    edit:command-history &dedup &newest-first &cmd-only |
    to-terminated "\x00" |
    try {
      fzf --no-sort --read0 --layout=reverse --ansi --color=light --exact ^
        --no-multi --query=$edit:current-command | slurp
    } catch {
      # If the user presses [Escape] to cancel the fzf operation it will exit
      # with a non-zero status. Ignore that we ran this function in that case.
      return
    }
  )
  set edit:current-command = (str:trim-right $new-cmd " \n")
}

set edit:insert:binding[Ctrl-R] = {|| history >/dev/tty 2>&1 }
```
