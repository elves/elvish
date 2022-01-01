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

# Commit History

These files used to live in a
[separate repository](https://github.com/elves/elvish.io). However, because
@xiaq did not merge the repositories in the correct way (he simply copied all
the files), the commit history is lost. Please see that repository for a full
list of contributors.
