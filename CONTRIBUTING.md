# Contributor's Manual

## Human communication

The project lead is @xiaq, who is reachable in the user group most of the time.

If you intend to make user-visible changes to Elvish's behavior, it is good idea
to talk to him first; this will make it easier to review your changes.

On the other hand, if you find it easier to express your thoughts directly in
code, it is also completely fine to directly send a pull request, as long as you
don't mind the risk of the PR being rejected due to lack of prior discussion.

## Testing changes

Write comprehensive unit tests for your code, and make sure that existing tests
are passing. Tests are run on CI automatically for PRs; you can also run
`make test` in the repo root yourself.

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

Reference docs are interspersed in Go sources as comments blocks whose first
line starts with `//elvdoc` (and are hence called _elvdocs_). They can use
[Github Flavored Markdown](https://github.github.com/gfm/).

Elvdocs for functions look like the following:

````go
//elvdoc:fn name-of-fn
//
// ```elvish
// name-of-fn $arg &opt=default
// ```
//
// Does something.
//
// Example:
//
// ```elvish-transcript
// ~> name-of-fn something
// ▶ some-value-output
// ```

func nameOfFn() { ... }
````

Generally, elvdocs for functions have the following structure:

-   A line starting with `//elvdoc:fn`, followed by the name of the function.
    Note that there should be no space after `//`, unlike all the other lines.

-   An `elvish` code block describing the signature of the function, following
    the convention [here](website/ref/builtin.md#usage-notation).

-   Description of the function, which can be one or more paragraphs. The first
    sentence of the description should start with a verb in 3rd person singular
    (i.e. ending with a "s"), as if there is an implicit subject "this
    function".

-   One or more `elvish-transcript` code blocks showing example usages, which
    are transcripts of actual REPL input and output. Transcripts must use the
    default prompt `~>` and default value output indicator `▶`. You can use
    `elvish -norc` if you have customized either in your `rc.elv`.

Place the comment block before the implementation of the function. If the
function has no implementation (e.g. it is a simple wrapper of a function from
the Go standard library), place it before the top-level declaration of the
namespace.

Similarly, reference docs for variables start with `//elvdoc:var`:

```go
//elvdoc:var name-of-var
//
// Something.
```

Variables do not have signatures, and are described using a noun phrase.
Examples are not always needed; if they are, they can be given in the same
format as examples for functions.

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

## Formatting source files

Install [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to
format Go files, and [prettier](https://prettier.io/) to format Markdown files.

```sh
go install golang.org/x/tools/cmd/goimports@latest
npm install --global prettier@2.3.1
```

Once you have installed the tools, use `make style` to format Go and Markdown
files. If you prefer, you can also configure your editor to run these commands
automatically when saving Go or Markdown sources.

Use `make checkstyle` to check if all Go and Markdown files are properly
formatted.

## Linting

Install [staticcheck](https://staticcheck.io):

```sh
go install honnef.co/go/tools/cmd/staticcheck@2021.1
```

The other linter Elvish uses is the standard `go vet` command. Elvish doesn't
use golint since it is
[deprecated and frozen](https://github.com/golang/go/issues/38968).

Use `make lint` to run `staticcheck` and `go vet`.

## Checking spelling

Install [codespell] to check spelling:

```sh
pip install --user codespell==2.1.0
```

Use `make codespell` to run it.

## Licensing

By contributing, you agree to license your code under the same license as
existing source code of elvish. See the LICENSE file.
