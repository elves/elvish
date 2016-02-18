# Notes for Contributors

## Testing

Always run unit tests before committing. `make` will take care of this.

## Generated files

Some files are generated from other files. They should be commmited into the repository for this package to be go-getable. Run `go generate ./...` to regenerate them in case you modified the source.

## Formatting the Code

Always format the code with `goimports` before committing. Run `go get golang.org/x/tools/cmd/goimports` to install `goimports`, and `goimports -w .` to format all golang sources.

To automate this you can set up a `goimports` filter for Git by putting this in `~/.gitconfig`:

    [filter "goimports"]
        clean = goimports
        smudge = cat

Git will then always run `goimports` for you before comitting, since `.gitattributes` in this repository refers to this filter. More about Git attributes and filters [here](https://www.kernel.org/pub/software/scm/git/docs/gitattributes.html).

## Licensing

By contributing, you agree to license your code under the same license as existing source code of elvish. See the LICENSE file.
