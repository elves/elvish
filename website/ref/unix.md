<!-- toc -->

# Introduction

The `unix:` module provides access to features that only make sense on UNIX
compatible operating systems such as Linux, FreeBSD, macOS, etc. On non-UNIX
operating systems, such as MS Windows, this namespace does not exist and
`use unix` will fail. See also the
[`$platform:is-unix`](platform.html#platformis-unix) variable which can be used
to determine if this namespace is usable.

@elvdoc -ns unix: -dir ../pkg/eval/unix
