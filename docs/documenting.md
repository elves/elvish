# Documenting changes

Always document user-visible changes.

## Release notes

Add a brief list item to the release note of the next release, in the
appropriate section. You can find the document at the root of the repo (called
`$version-release-notes.md`).

## Reference docs

Reference docs are written as "elvdocs", comment blocks before unindented `fn`
or `var` declarations in Elvish files. A
[large subset](https://pkg.go.dev/src.elv.sh/pkg/md@master) of
[CommonMark](https://commonmark.org) is supported. Examples:

````elvish
# Does something.
#
# Examples:
#
# ```elvish-transcript
# ~> foo
# some output
# ```
fn foo {|a b c| }

# Some variable.
var bar
````

Most of Elvish's builtin modules are implemented in Go, not Elvish. For those
modules, put dummy declarations in `.d.elv` files (`d` for "declaration"). For
example, elvdocs for functions implemented in `builtin_fn_num.go` go in
`builtin_fn_num.d.elv`.

For a comment block to be considered an elvdoc, it has to be continuous, and
each line should either be just `#` or start with `#` and a space.

Style guides for elvdocs for functions:

-   The first sentence should start with a verb in 3rd person singular (i.e.
    ending with a "s"), as if there is an implicit subject "this function".

-   The end of the elvdoc should show or more `elvish-transcript` code blocks
    showing example usages, which are transcripts of actual REPL input and
    output. Transcripts must use the default prompt `~>` and default value
    output indicator `▶`. You can use `elvish -norc` if you have customized
    either in your [`rc.elv`](https://elv.sh/ref/command.html#rc-file).

It is quite common for elvdocs to link to other elvdocs, and Elvish's website
toolchain provides special support for that. If a link has a single code span
and an empty target, it gets rewritten to a link to an elvdoc section. For
example, ``[`put`]()`` will get rewritten to ``[`put`](builtin.html#put)``, or
just ``[`put`](#put)`` within the documentation for the builtin module.

## Comment for unexported Go types and functions

In the doc comment for exported types and functions, it's customary to use the
symbol itself as the first word of the comment. For unexported types and
functions, this becomes a bit awkward as their names don't start with a capital
letter, so don't repeat the symbol. Examples:

```go
// Foo does foo.
func Foo() { }

// Does foo.
func foo() { }
```
