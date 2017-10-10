# Notes for Contributors

## Code

1.  Always run unit tests before committing. `make` will take care of this.

2.  Some files are generated from other files. They should be committed into the repository for this package to be go-getable. Run `go generate ./...` to regenerate them in case you modified the source.

3.  Always format the code with `goimports` before committing. Run `go get golang.org/x/tools/cmd/goimports` to install `goimports`, and `goimports -w .` to format all golang sources.

    To automate this you can set up a `goimports` filter for Git by putting this in `~/.gitconfig`:

        [filter "goimports"]
            clean = goimports
            smudge = cat

    Git will then always run `goimports` for you before committing, since `.gitattributes` in this repository refers to this filter. More about Git attributes and filters [here](https://www.kernel.org/pub/software/scm/git/docs/gitattributes.html).

## Human Communication

If you are making significant changes or any user-visible changes (e.g. changes to the language, the UI or the elv script API), please discuss on the developer channel before starting to work.

## Licensing

By contributing, you agree to license your code under the same license as existing source code of elvish. See the LICENSE file.
