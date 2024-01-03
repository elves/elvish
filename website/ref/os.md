<!-- toc -->

@module os

# Introduction

The `os:` module provides access to operating system functionality. The
interface is intended to be uniform across operating systems.

The [builtin module](builtin.html) also contains some operating system
utilities. The [`unix:` module](unix.html) contains utilities that are specific
to UNIX operating systems.

Function usages are given in the same format as in the reference doc for the
[builtin module](builtin.html).

## The `&follow-symlink` option {#follow-symlink}

Some commands take an `&follow-symlink` option, which controls the behavior of
the command when the final element of the path is a symbolic link:

-   When the option is false (usually the default), the command operates on the
    symbolic link file itself.

-   When the option is true, the commands operates on the file the symbolic link
    points to.

As an example, when `l` is a symbolic link to a directory:

-   `os:is-dir &follow-symlink=$false l` outputs `$false`, since the symbolic
    link file itself is not a directory.

-   `os:is-dir &follow-symlink=$true l` outputs `$true`, since the symlink
    points to a directory.
