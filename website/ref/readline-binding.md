<!-- toc -->

@module readline-binding

# Introduction

The `readline-binding` module provides readline-like key bindings, such as
binding <span class="key">Ctrl-A</span> to move the cursor to the start of the
line. To use, put the following in `~/.elvish/rc.elv`:

```elvish
use readline-binding
```

Note that this will override some of the standard bindings. For example, <span
class="key">Ctrl-L</span> will be bound to a function that clears the terminal
screen rather than start location mode.

See the
[source code](https://github.com/elves/elvish/blob/master/pkg/mods/bundled/readline-binding.elv.go)
for details.
