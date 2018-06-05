This directory contains source for Elvish's official website. They are
converted to HTML with a custom toolchain with the following dependencies:

*   GNU Make (any "reasonably modern" version should do).

*   pandoc 2.2.1 (other versions in the 2.x series might also work).

*   A Go toolchain, for building [genblog](https://github.com/xiaq/genblog)
    and some custom preprocessors in the `tools` directory.

The documents are written in GitHub-flavored markdown sprinkled with some HTML
and custom macros. Most of them can be viewed directly in GitHub.

# History

These files used to live in a [separate
repository](https://github.com/elves/elvish.io). However, because @xiaq did
not merge the repositories in the correct way (he simply copied all the
files), the commit history is lost. Please see that repository for a full list
of contributors.
