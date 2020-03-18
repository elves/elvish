# Contributor's Manual

## Human communication

The project lead is @xiaq, who is reachable in the user group most of the time.
If you intend to make user-visible changes to Elvish's behavior, it good idea to
talk to him first; this will make it easier to review your changes.

## Unit tests

Write comprehensive unit tests for your code, and make sure that existing tests
are passing. Running `make` in the repo root will run all unit tests.

Most of the Elvish codebase has good testing utilities today, and some also have
established testing patterns. Before writing unit tests, read a few existing
tests in the package you are changing, and follow the existing patterns. Some
packages unfortunately have two (and hopefully no more than that) competing
patterns. When in doubt, ask the project lead.

Still, some part of the codebase is poorly tested, and may even be outright
untestable. In that case, also discuss to the project lead.

Most of the Elvish codebase has decent unit test coverage. When contributing to
Elvish,

## Documenting builtin functions and variables

Documentation for builtin functions and variables are written in the Go sources
as comments.

Document function using the following format:

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
````

Note the following:

-   There is no space before `elvdoc` of the first line.

-   The signature of the function follows the convention
    [here](website/ref/builtin.md#usage-notation).

-   The first sentence of the description should start with a verb in 3rd person
    singular (i.e. ending with a "s"), as if there is an implicit subject "this
    function".

-   Examples are best illustrated by a transcript of actual REPL input and
    output. Always use the default prompt `~>` and default value output
    indicator `▶`. If you have customized those, it's left as an exercise to
    define a key binding in `rc.elv` for switching between the default and your
    customized values.

Document variables using the following format:

```go
//elvdoc:var name-of-var
//
// Something.
```

Variables do not have signatures, and can be described using a noun phrase.
Examples are not always needed; if they are, they can be given in the same
format as examples for functions above.

## Generating code

Elvish uses generated code in a few places. As is the usual case with Go
projects, they are committed into the repo, and if you change the input of a
generated file you should re-generate it.

Use the standard command, `go generate ./...` to regenerate all files.

Dependencies of the generation rules:

-   The `stringer` tool: Install with
    `go get -u golang.org/x/tools/cmd/stringer`;

-   An installed `elvish` in your PATH;

-   Python 2.7 at `/usr/bin/python2.7`.

    NOTE: Python scripts should be rewritten in either Go or Elvish, but we
    still have some.

## Formatting source files

Format Go code with
[goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports):

```sh
go get golang.org/x/tools/cmd/goimports # Install
goimports -w . # Format Go files
```

Format Markdown files with [prettier](https://prettier.io/):

```sh
npm install --global prettier # Install
prettier --tab-width 4 --prose-wrap always --write *.md # Format Markdown files
```

Refer to the documentation of your editor to run these commands automatically
when saving Go or Markdown sources.

## Licensing

By contributing, you agree to license your code under the same license as
existing source code of elvish. See the LICENSE file.
