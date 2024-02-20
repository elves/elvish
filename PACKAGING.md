# Packager's Manual

The main package of Elvish is `cmd/elvish`, and you can build it like any other
Go application.

## Enhancing version information

There are two variables that may be set during compilation using linker flags to
enhance Elvish's version information (they don't affect any other aspect of
Elvish's behavior):

-   `src.elv.sh/pkg/buildinfo.BuildVariant` can be set to identify the
    distribution you're building for. You are recommended to set this variable.

-   `src.elv.sh/pkg/buildinfo.VCSOverride` can be set to supply VCS metadata for
    development builds. This is only needed in the rare occasion where:

    -   You are packaging development builds

    -   Go's mechanism to store VCS metadata doesn't work for your build
        environment

See [godoc](https://pkg.go.dev/src.elv.sh@master/pkg/buildinfo#pkg-variables)
for detailed information on the semantics of these two variables and how to set
them.

**Note**: The guidance here applies to the current development version and
release versions starting from 0.19.0. The details for earlier versions are
different. If your build script has `-ldflags '-X $symbol=$value'` where
`$symbol` is not documented here, those flags no longer do anything and should
be removed.
