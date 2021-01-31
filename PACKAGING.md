# Packager's Manual

**Note**: The guidance here applies to the current development version and
release versions starting from v0.15.0. The details for earlier versions are
different.

Elvish is a normal Go application, and doesn't require any special attention.
Build the main package of `cmd/elvish`, and you should get a fully working
binary.

If you don't care about accurate version information or reproducible builds, you
can now stop reading. If you do, there is a small amount of extra work to get
them.

## Accurate version information

The `pkg/buildinfo` package contains a constant, `Version`, and a variable,
`VersionSuffix`, which are concatenated to form the full version used in the
output of `elvish -version` and `elvish -buildinfo`. Their values are set as
follows:

-   At release tags, `Version` contains the version of the release, which is
    identical to the tag name. `VersionSuffix` is empty.

-   At development commits, `Version` contains the version of the next release.
    `VersionSuffix` is set to `-dev.unknown`.

The `VersionSuffix` variable can be overriden at build time, by passing
`-ldflags "-X src.elv.sh/pkg/buildinfo.VersionSuffix=-foobar"` to `go build`,
`go install` or `go get`. This is necessary in several scenarios, which are
documented below.

### Packaging release versions

If you are not applying any patches, there is nothing to do. The default value
of `VersionSuffix`, which is empty, suffices.

If you have applied any patches, you **must** override `VersionSuffix` with a
string that starts with `+` and can uniquely identify your patch. For official
Linux distribution builds, this should identify your distribution, plus the
version of the patch. Example:

```sh
go build -ldflags "-X src.elv.sh/pkg/buildinfo.VersionSuffix=+deb1" ./cmd/elvish
```

### Packaging development builds

If you are packaging development builds, the default value of `VersionSuffix`,
which is `-dev.unknown`, is likely not good enough, as it does not identify the
commit Elvish is built from.

You should override `VersionSuffix` with `-dev.$commit_hash`, where
`$commit_hash` is the full commit hash, which can be obtained with
`git rev-parse HEAD`. Example:

```sh
go build -ldflags \
  "-X src.elv.sh/pkg/buildinfo.VersionSuffix=-dev.$(git rev-parse HEAD)" \
  ./cmd/elvish
```

If you have applied any patches that is not committed as a Git commit, you
should also append a string that starts withs `+` and can uniquely identify your
patch.

## Reproducible builds

The idea of
[reproducible build](https://en.wikipedia.org/wiki/Reproducible_builds) is that
an Elvish binary from two different sources should be bit-to-bit identical, as
long as they are built from the same version of the source code using the same
version of the Go compiler.

To make reproducible builds, you must do the following:

-   Pass `-trimpath` to the Go compiler.

-   For Linux and Windows, also pass `-buildmode=pie` to the Go compiler.

-   Disable cgo by setting the `CGO_ENABLED` environment variable to 0.

-   Follow the requirements above for putting
    [accurate version information](#accurate-version-information) into the
    binary, so that the user is able to uniquely identify the build by running
    `elvish -version`.

    The recommendation for how to set `VersionSuffix` when
    [packaging development builds](#packaging-development-builds) becomes hard
    requirements when packaging reproducible builds.

    In addition, if your distribution uses a patched version of the Go compiler
    that changes its output, or if the build command uses any additional flags
    (either via the command line or via any environment variables), you must
    treat this as a patch on Elvish itself, and supply a version suffix
    accordingly.

If you follow these requirements when building Elvish, you can mark the build as
a reproducible one by overriding `src.elv.sh/pkg/buildinfo.Reproducible` to
`"true"`.

Example when building a release version without any patches for Linux or
Windows:

```sh
go build -buildmode=pie -trimpath \
  -ldflags "-X src.elv.sh/pkg/buildinfo.Reproducible=true" \
  ./cmd/elvish
```

Example when building a development version with a patch for Linux or Windows:

```sh
go build -buildmode=pie -trimpath \
  -ldflags "-X src.elv.sh/pkg/buildinfo.VersionSuffix=-dev.$(git rev-parse HEAD)+deb0 \
            -X src.elv.sh/pkg/buildinfo.Reproducible=true" \
  ./cmd/elvish
```
