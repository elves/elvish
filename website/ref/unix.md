<!-- toc -->

@module unix

# Introduction

The `unix:` module provides access to features that only make sense on UNIX-like
operating systems, such as Linux, FreeBSD, and macOS.

On non-UNIX operating systems, such as MS Windows, this namespace does not exist
and `use unix` will fail. Use the
[`$platform:is-unix`](platform.html#$platform:is-unix) variable to determine if
this namespace is usable.
