This directory contains source for Elvish's official website.

The documents are written in GitHub-flavored markdown sprinkled with some HTML
and custom macros. Most of them can be viewed directly in GitHub; notable
exceptions are the homepage (`home.md`) and the download page
(`get/prelude.md`).

# Building

The website is a collection of static HTML files, built from Markdown files with
a custom toolchain. You need the following software to build it:

-   Go, with the same version requirement as Elvish itself.

-   Pandoc 2.2.1 (other versions in the 2.x series might also work).

-   GNU Make (any "reasonably modern" version should do).

To build the website, just run `make`. The built website is in the `_dst`
directory. You can then open `_dst/index.html` or run an HTTP server within
`_dst` to preview.

**NOTE**: Although the website degrades gracefully when JavaScript is disabled,
local viewing works best with JavaScript enabled. This is because relative paths
like `./get` will cause the browser to open the corresponding directory, instead
of the `index.html` file under it, and we use JavaScript to patch such URLs
dynamically.

## Building the docset

Building the docset requires the following additional dependencies:

-   Python 3 with Beautiful Soup 4 (install with `pip install bs4`).

-   SQLite3 CLI.

To build the docset, run `make docset`. The generated docset is in
`Elvish.docset`.

# Ttyshots

Some of the pages include "ttyshots" that show the content of Elvish sessions.
They are HTML files with terminal attributes converted to CSS classes, generated
from corresponding instruction files. By convention, the instruction files have
names ending in `.ttyshot`, and the generated HTML files have names ending in
`.ttyshot.html`.

The generation process depends on [`tmux`](https://github.com/tmux/tmux) and a
built `elvish` in `PATH`. Windows is not supported.

## Instruction syntax

Each line in a ttyshot instruction file is one of the following:

-   `#prompt` instructs waiting for a new shell prompt.

-   `#`_`command`_, where `command` is a string that does **not** start with a
    space and is not `prompt`, is a command sent to `tmux`. The most useful one
    (and only one being used now) is `send-keys`.

-   Anything else is treated as text that should be sent directly to the Elvish
    prompt.

For example, the following instructions runs `cd /tmp`, waits for the next
prompt, and sends Ctrl-N to trigger navigation mode:

```
cd /tmp
#prompt
#send-keys ctrl-L
```

## Generating ttyshots

Unlike other generated website artifacts, generated ttyshots are committed into
the repository, and the `Makefile` rule to generate them is disabled by default.
This is because the process to generate ttyshots is relatively slow and may have
network dependencies.

To turn on ttyshot generation, pass `TTYSHOT=1` to `make`. For example, to
generate a single ttyshot, run `make TTYSHOT=1 foo.ttyshot.html`. To build the
website with ttyshot generation enabled, run `make TTYSHOT=1`.

The first time you generate ttyshots, `make` will build the `ttyshot` tool, and
regenerate all ttyshots. Subsequent runs will only regenerate ttyshots whose
instruction files have changed.

# Commit History

These files used to live in a
[separate repository](https://github.com/elves/elvish.io). However, because
@xiaq did not merge the repositories in the correct way (he simply copied all
the files), the commit history is lost. Please see that repository for a full
list of contributors.
