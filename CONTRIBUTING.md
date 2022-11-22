# Contributor's Manual

## Human communication

The project lead is @xiaq, who is reachable in the user group most of the time.

If you intend to make user-visible changes to Elvish's behavior, it is good idea
to talk to him first; this will make it easier to review your changes.

On the other hand, if you find it easier to express your thoughts directly in
code, it is also completely fine to directly send a pull request, as long as you
don't mind the risk of the PR being rejected due to lack of prior discussion.

## Using development scripts

The [`Makefile`](Makefile) contains targets encapsulating some common workflows.
They are not necessary for developing Elvish, but can save you a few keystrokes.
GNU Make is required.

The [`tools`](tools) directory contains scripts too complex to fit in the
`Makefile`. Among them, [`tools/pre-push`](tools/pre-push) can be used as a Git
hook, and covers all the CI checks that can be run from your local environment.

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

## Code hygiene

Some basic aspects of code hygiene are checked in the CI.

### Formatting

Install [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to
format Go files.

```sh
go install golang.org/x/tools/cmd/goimports@latest
```

The Markdown formatter [elvmdfmt](cmd/elvmdfmt) lives inside this repo and does
not need to be installed.

Once you have installed the tools, use `make style` to format Go and Markdown
files, or `make checkstyle` to check if all Go and Markdown files are properly
formatted.

#### Formatting on save

The Go plugins of most popular editors already support formatting Go files
automatically on save; consult the documentation of the plugin you use.

To format Markdown files automatically on save, configure your editor to run the
following command when saving Markdown files:

```sh
go run src.elv.sh/cmd/elvmdfmt -width 80 -w $filename
```

**Note**: Using `go run` ensures that you are always using the `elvmdfmt`
implementation in the repo, but it incurs a small performance penalty since the
Go toolchain does not cache binary files. If this is a problem (for example, if
your editor runs the command synchronously), you can speed up the command by
installing `src.elv.sh/cmd/elvmdfmt` and using the installed `elvmdfmt`.
However, if you do this, you must re-install `elvmdfmt` whenever there is a
change in its implementation that impacts the output.

You'll also want to configure this command to only run inside the Elvish repo,
since `elvmdfmt` is tailored to Markdown files in this repo and may not work
well for other Markdown files.

If you use VS Code, you can install the
[Run on Save](https://marketplace.visualstudio.com/items?itemName=emeraldwalk.RunOnSave)
extension and add the following to the workspace (not user) `settings.json`
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

### Linting

Install [staticcheck](https://staticcheck.io):

```sh
go install honnef.co/go/tools/cmd/staticcheck@v0.3.2
```

The other linter Elvish uses is the standard `go vet` command. Elvish doesn't
use golint since it is
[deprecated and frozen](https://github.com/golang/go/issues/38968).

Use `make lint` to run `staticcheck` and `go vet`.

### Spell checking

Install [codespell](https://github.com/codespell-project/codespell) to check
spelling:

```sh
pip install --user codespell==2.2.1
```

Use `make codespell` to run it.

### Running all checks

Use this command to run all checks:

```sh
make test checkstyle lint codespell
```

You can put this in `.git/hooks/pre-push` to ensure that your published commits
pass all the checks.

## Licensing

By contributing, you agree to license your code under the same license as
existing source code of elvish. See the LICENSE file.
