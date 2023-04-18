<!-- toc -->

@module os

# Introduction

The `os:` module provides commands for performing operating system specific
functions such as creating directories and removing files. The commands in this
module are, for the most part, OS agnostic. That is, commands such as
`os:remove` can be expected to behave more or less the same regardless of the
the platform Elvish is running on. Where that is not true it will be noted in
the documentation of the command.

Function usages are given in the same format as in the reference doc for the
[builtin module](builtin.html).
