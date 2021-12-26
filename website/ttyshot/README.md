# What this directory contains

This directory contains "ttyshots" that represent the final state of a set of
Elvish interactive shell interactions. You will need to have the
[`tmux`](https://github.com/tmux/tmux) command installed to create these images.
The process for generating a ttyshot does not require a real terminal unless you
want to examine the "raw" image. This means ttyshots can be generated in a CI/CD
workflow using nothing more than the "spec" files.

# How to create a ttyshot

To create a ttyshot use the following procedure from the project root dir:

1.  Create the ttyshot program: `make -C website tools/ttyshot.bin`

1.  Create or modify a ttyshot specification (a ".spec" file) in the
    `website/ttyshot` directory or subdirectory. See below for what can appear
    in a ".spec" file.

1.  Create the ttyshot; e.g.,
    `website/tools/ttyshot.bin website/ttyshot/pipelines.spec`.

1.  Review the results; e.g., `cat website/ttyshot/pipelines.raw`.

1.  Add a `@ttyshot` directive to the appropriate document (e.g.,
    `website/home.md` or `website/learn/fundamentals.md`) if adding a new
    ttyshot.

You can easily refresh all the ttyshots by running this:

```
for f [website/ttyshot/**.spec] { put $f; website/tools/ttyshot.bin $f }
```

# Content of a ttyshot specification

A ttyshot specification consists of two types of lines: plain text to be sent to
the Elvish shell as if a human had typed the text, and directives that begin
with `//`. Empty lines are ignored. The available directives are:

-   `//prompt`: Wait for a new shell prompt. The process of converting a
    sequence of commands to a ttyshot implicitly waits for the first prompt to
    appear so don't begin your specification with this directive.

-   `//no-enter`: Disable the implicit Enter normally sent after each line of
    plain text.

-   `//enter`: Enable an implicit Enter after each line of plain text and send
    an Enter.

-   `//sleep d`: Pause for the specified duration in seconds. For example:
    `//sleep 1`. No unit suffix should be present since only (fractional)
    seconds are allowed. The use of this directive shouldn't be necessary. There
    is an implicit sleep at the end of a ttyshot specification before capturing
    the ttyshot image to give the Elvish shell time to stabilize its output;
    e.g., when displaying a navigation view.

-   `//alt x`: Simulate an Alt sequence. That is, send an Escape followed by the
    specified text.

-   `//ctrl n`: Send a Ctrl char version of `n`; e.g., `//ctrl L`.

-   `//tab`: Send a Tab; i.e., `//ctrl I`.

-   `//up`: Send an Up-arrow key sequence.

-   `//down`: Send an Down-arrow key sequence.

-   `//left`: Send an Left-arrow key sequence.

-   `//right`: Send an Right-arrow key sequence.

-   `//wait-for-str string`: Wait for the literal string `string` to appear in
    the output.

-   `//wait-for-re regexp`: Wait for text matching the regexp to appear in the
    output.

## Example ttyshot specification

The following specification simulates a user pressing Ctrl-N to enter navigation
mode. Followed by Ctrl-F and the text `pkg` to select that directory. Followed
by navigating into and out of that directory.

```
//no-enter
//ctrl N
//down
//ctrl F
pkg
//right
sys
//right
```
