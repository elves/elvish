# Source for Elvish's website

This directory contains source for Elvish's official website.

The documents are written in [CommonMark](https://commonmark.org) sprinkled with
some HTML and custom macros. Most of them can be viewed directly in GitHub;
notable exceptions are the homepage (`home.md`) and the download page
(`get/prelude.md`).

## Building

The website is a collection of static HTML files, built from Markdown files with
a custom toolchain. You need the following software to build it:

-   Go, with the same version requirement as Elvish itself.

-   GNU Make (any "reasonably modern" version should do).

To build the website, just run `make`. The built website is in the `_dst`
directory. You can then open `_dst/index.html` or run an HTTP server within
`_dst` to preview.

**NOTE**: Although the website degrades gracefully when JavaScript is disabled,
local viewing works best with JavaScript enabled. This is because relative paths
like `./get` will cause the browser to open the corresponding directory, instead
of the `index.html` file under it, and we use JavaScript to patch such URLs
dynamically.

### Additional workflows

-   Run `make check-rellinks` to ensure that relative links between pages are
    valid.

-   Run `make Elvish.docset` to build a docset containing all the reference
    docs. [Docset](https://kapeli.com/docsets) is a format for packaging
    documentation for offline consumption.

Both workflows use a Python script under the hood, and require Python 3 and
Beautiful Soup 4 (install with `pip install --user bs4`).

## Transcripts

Documents can contain **transcripts** of Elvish sessions, identified by the
language tag `elvish-transcript`. A simple example:

````markdown
```elvish-transcript
~> echo foo |
   str:to-upper (one)
▶ FOO
```
````

When the website is built, the toolchain will highlight the
`echo foo | str:to-upper (one)` part correctly as Elvish code.

To be exact, the toolchain uses the following heuristic to determine the range
of Elvish code:

-   It looks for what looks like a prompt, which starts with either `~` or `/`,
    ends with `>` and a space, with no spaces in between.

-   It then extends the range downwards, as long as the line starts with N
    whitespaces, where N is the length of the prompt (including the trailing
    space).

As long as you use Elvish's default prompt, you should be able to rely on this
heuristic.

## Ttyshots

Some of the pages include "ttyshots" that show the content of Elvish sessions.
They are HTML files with terminal attributes converted to CSS classes, generated
from corresponding instruction files. By convention, the instruction files have
names ending in `-ttyshot.elvts` (because they are syntactically Elvish
transcripts), and the generated HTML files have names ending in `-ttyshot.html`.

The generation process depends on [`tmux`](https://github.com/tmux/tmux) and a
built `elvish` in `PATH`. Windows is not supported.

### Instruction syntax

Ttyshot instruction files look like Elvish transcripts, with the following
differences:

-   It should not contain the output of commands. Anything that is not part of
    an input at a prompt causes a parse error.

-   If the Elvish code starts with `#` followed immediately by a letter, it is
    treated instead as a command to sent to `tmux`.

    The most useful one (and only one being used now) is `send-keys`.

For example, the following instructions runs `cd /tmp`, and sends Ctrl-N to
trigger navigation mode at the next prompt:

```elvish-transcript
~> cd /tmp
~> #send-keys C-N
```

### Generating ttyshots

Unlike other generated website artifacts, generated ttyshots are committed into
the repository, and the `Makefile` rule to generate them is disabled by default.
This is because the process to generate ttyshots is relatively slow and may have
network dependencies.

To turn on ttyshot generation, pass `TTYSHOT=1` to `make` (where `1` can be
replaced by any non-empty string). For example, to generate a single ttyshot,
run `make TTYSHOT=1 foo-ttyshot.html`. To build the website with ttyshot
generation enabled, run `make TTYSHOT=1`.

The first time you generate ttyshots, `make` will build the `ttyshot` tool, and
regenerate all ttyshots. Subsequent runs will only regenerate ttyshots whose
instruction files have changed.

## Commit History

These files used to live in a
[separate repository](https://github.com/elves/elvish.io). However, because
@xiaq did not merge the repositories in the correct way (he simply copied all
the files), the commit history is lost. Please see that repository for a full
list of contributors.
