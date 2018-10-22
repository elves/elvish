This directory contains source for Elvish's official website.

The documents are written in GitHub-flavored markdown sprinkled with some HTML
and custom macros. Most of them can be viewed directly in GitHub; notable
exceptions are the homepage (`home.md`) and the download page
(`download/prelude.md`).

# Building

The website is a purely static one. It is built with a custom toolchain with
the following dependencies:

*   GNU Make (any "reasonably modern" version should do).

*   Pandoc 2.2.1 (other versions in the 2.x series might also work).

*   A Go toolchain, for building [genblog](https://github.com/xiaq/genblog)
    and some custom preprocessors in the `tools` directory.

To build the website, just run `make`. The built website is in the `_dst`
directory; to preview, run an HTTP server within it.

Opening `_dst/index.html` almost works, except that browsers will show the
directories as file lists instead of using the `index.html` file, so if you
click e.g. "Learn" in the nav bar, you will need to manually click the
`index.html` within it.

# Commit History

These files used to live in a [separate
repository](https://github.com/elves/elvish.io). However, because @xiaq did
not merge the repositories in the correct way (he simply copied all the
files), the commit history is lost. Please see that repository for a full list
of contributors.
