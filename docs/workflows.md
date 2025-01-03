# Common development workflows

The [`Makefile`](Makefile) encapsulates common development workflows:

-   Use `make fmt` to [format files](#formatting-files).

-   Use `make test` to [run tests](./testing.md).

-   Use `make all-checks` or `make most-checks` to
    [run checks](#running-checks).

You can use the [`tools/pre-push`](../tools/pre-push) script as a Git hook,
which runs all the tests and checks (`make test all-checks`), among other
things.

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

The Markdown formatter is [`cmd/elvmdfmt`](../cmd/elvmdfmt), which lives inside
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
([`tools/check-gen.sh`](../tools/check-gen.sh)) that requires the Git repo to
have a clean working tree, so may not be convenient to use when you are working
on the source code. The `most-checks` target excludes that, so can be always be
used.

The checks depend on some external programs, which can be installed as follows:

<!-- Keep the versions of staticcheck and codespell in sync with .github/workflows/ci.yml -->

```sh
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@v0.5.1
pip install --user codespell==2.3.0
```

## Licensing

By contributing, you agree to license your code under the same license as
existing source code of elvish. See the LICENSE file.
