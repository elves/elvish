# Contributor's Manual

## Human communication

The only person with direct commit access is the project's founder @xiaq. If you
intend to make user-visible changes to Elvish's behavior (as opposed to fixing
typos and obvious bugs), it is good idea to talk to him first; this will make it
easier to review your changes. He should be reachable in the user group most of
the time.

On the other hand, if you find it easier to express your thoughts directly in
code, it is also completely fine to directly send a pull request, as long as you
don't mind the risk of the PR being rejected due to lack of prior discussion.

## Development workflows

The [`Makefile`](Makefile) encapsulates common development workflows:

-   Use `make fmt` to [format files](#formatting-files).

-   Use `make test` to [run tests](#testing-changes).

-   Use `make all-checks` or `make most-checks` to
    [run checks](#running-checks).

You can use the [`tools/pre-push`](tools/pre-push) script as a Git hook, which
runs all the tests and checks (`make test all-checks`), among other things.

The same tests and checks are also run by Elvish's CI environments, so running
them locally before pushing minimizes the chance of CI errors. (The CI
environments run the tests on multiple platforms, so CI errors can still happen
if you break some tests for a different platform.)

## Formatting files

Use `make fmt` to format Go and Markdown files in the repo.

### Formatting Go files on save

The Go plugins of most popular editors already support formatting Go files
automatically on save; consult the documentation of the plugin you use.

### Formatting Markdown files on save

The Markdown formatter is [`cmd/elvmdfmt`](cmd/elvmdfmt), which lives inside
this repo. Run it like this:

```sh
go run src.elv.sh/cmd/elvmdfmt -width 80 -w $filename
```

To format Markdown files automatically on save, configure your editor to run the
command above when saving Markdown files. You'll also want to configure this
command to only run inside the Elvish repo, since `elvmdfmt` is tailored to
Markdown files in this repo and may not work well for other Markdown files.

If you use VS Code, install the
[Run on Save](https://marketplace.visualstudio.com/items?itemName=emeraldwalk.RunOnSave)
extension and add the following to the workspace (**not** user) `settings.json`
file:

```json
"emeraldwalk.runonsave": {
    "commands": [
        {
            "match": "\\.md$",
            "cmd": "go run src.elv.sh/cmd/elvmdfmt -width 80 -w ${file}"
        }
    ]
}
```

**Note**: Using `go run` ensures that you are always using the `elvmdfmt`
implementation in the repo, but it incurs a small performance penalty since the
Go toolchain does not cache binary files and has to rebuild it every time. If
this is a problem (for example, if your editor runs the command synchronously),
you can speed up the command by installing `src.elv.sh/cmd/elvmdfmt` and using
the installed `elvmdfmt`. However, if you do this, you must re-install
`elvmdfmt` whenever there is a change in its implementation that impacts the
output.

## Testing changes

Write comprehensive unit tests for your code, and make sure that existing tests
are passing. Run tests with `make test`.

Respect established patterns of how unit tests are written. Some packages
unfortunately have competing patterns, which usually reflects a still-evolving
idea of how to best test the code. Worse, parts of the codebase are poorly
tested, or even untestable. In either case, discuss with the project lead on the
best way forward.

### ELVISH_TEST_TIME_SCALE

Some unit tests depend on time thresholds. The default values of these time
thresholds are suitable for a reasonably powerful laptop, but on
resource-constraint environments (virtual machines, embedded systems) they might
not be enough.

Set the `ELVISH_TEST_TIME_SCALE` environment variable to a number greater than 1
to scale up the time thresholds used in tests. The CI environments use
`ELVISH_TEST_TIME_SCALE = 10`.

## Documenting changes

Always document user-visible changes.

### Release notes

Add a brief list item to the release note of the next release, in the
appropriate section. You can find the document at the root of the repo (called
`$version-release-notes.md`).

### Reference docs

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

### Comment for unexported Go types and functions

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

## Generating code

Elvish uses generated code in a few places. As is the usual case with Go
projects, they are committed into the repo, and if you change the input of a
generated file you should re-generate it.

Use the standard command, `go generate ./...` to regenerate all files.

Some of the generation rules depend on the `stringer` tool. Install with
`go install golang.org/x/tools/cmd/stringer@latest`.

## Running checks

There are some checks on the source code that can be run with `make all-checks`
or `make most-checks`. The difference is that `all-checks` includes a check
([`tools/check-gen.sh`](tools/check-gen.sh)) that requires the Git repo to have
a clean working tree, so may not be convenient to use when you are working on
the source code. The `most-checks` target excludes that, so can be always be
used.

The checks depend on some external programs, which can be installed as follows:

<!-- Keep the versions of staticcheck and codespell in sync with .github/workflows/ci.yml -->

```sh
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@v0.4.6
pip install --user codespell==2.2.6
```

## Licensing

By contributing, you agree to license your code under the same license as
existing source code of elvish. See the LICENSE file.
