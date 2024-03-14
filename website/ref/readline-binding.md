<!-- toc -->

@module readline-binding

# Introduction

The `readline-binding` module provides GNU readline-like key bindings, such as
binding <kbd>Ctrl-A</kbd> to move the cursor to the start of the line. GNU
readline bindings are the default for shells such as Bash. So if you are
migrating from Bash to Elvish you probably want to add the following to your
[`rc.elv`](command.html#rc-file):

```elvish
use readline-binding
```

Note that this will override some of the standard bindings. For example,
<kbd>Ctrl-L</kbd> will be bound to a function that clears the terminal screen
rather than start location mode. The standard bindings are usually relocated to
use Alt as the modifier -- the location mode is bound to <kbd>Alt-L</kbd> for
example.

See the
[source code](https://src.elv.sh/pkg/mods/readline-binding/readline-binding.elv)
for details.
