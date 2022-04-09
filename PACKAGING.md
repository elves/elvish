# Packager's Manual

The main package of Elvish is `cmd/elvish`, and you can build it like any other
Go application. None of the instructions below are strictly required.

**Note**: The guidance here applies to the current development version and
release versions starting from 0.19.0. The details for earlier versions are
different.

## Identifying the build variant

You are encouraged to identify your build by overriding
`src.elv.sh/pkg/buildinfo.BuildVariant` with something that identifies the
distribution you are building for, and any patch level you have applied for
Elvish. This will allow Elvish developers to easily identify any
distribution-specific issue:

```sh
go build -ldflags '-X src.elv.sh/pkg/buildinfo.BuildVariant=deb1' ./cmd/elvish
```

### Official builds

A special build variant is `official`. This variant has a special meaning: the
binary must be bit-by-bit identical to the official binaries, linked from
https://elv.sh/get.

The official binaries are built using the `tools/buildall.sh` script in the Git
repo, using the docker image defined in https://github.com/elves/up. If you can
fully mirror the environment **and** verify that the resulting binary is
bit-by-bit identical to the official one, you can identify your build as
`official`.

**Important**: Reproducing the official binaries is not a one-off investment,
but an ongoing commitment: whenever the configuration for the official builds
changes, you must update your build setup accordingly. You must always verify
that your binaries are identical to official ones. As an example of how this can
be done, the
[check website](https://github.com/elves/elvish/blob/master/.github/workflows/check_website.yml)
workflow builds binaries for commits on the `master` branch and compare them
with official ones. If your build setup is technically reproducible, but you are
ready to ensure it's always identical to official binaries, you can always use a
distribution-specific variant, such as `deb1-reproducible`.

## Supplying VCS information for development builds

**Note**: This section is only relevant when building **development commits**,
which most distributions do not provide. Release commits always have their
version hardcoded in the code.

When Elvish is built from a development branch, it will try to figure out its
version from the VCS information Go compiler encoded. When that works,
`elvish -version` will output something like this:

```
0.19.0-dev.0.20220320172241-5dc8c02a32cf
```

The version string follows the syntax of
[Go module pseudo-version](https://go.dev/ref/mod#pseudo-versions), and consists
of the following parts:

-   `0.19.0-dev` identifies that this is a development build **before** the
    0.19.0 release.

-   `.0` indicates that this is a pseudo-version, instead of a real version.

-   `20220320172241` identifies the commit's creation time, in UTC.

-   `5dc8c02a32cf` is the 12-character prefix of the commit hash.

If that doesn't work for your build environment, the output of `elvish -version`
will instead be:

```
0.19.0-dev.unknown
```

If your build environment has the required information to build the
pseudo-version string, you can supply it by overriding
`src.elv.sh/pkg/buildinfo.VCSOverride` with the last two parts of the version
string, commit's creation time and the 12-character prefix of the commit hash:

```sh
go build -ldflags '-X src.elv.sh/pkg/buildinfo.VCSOverride=20220320172241-5dc8c02a32cf' ./cmd/elvish
```
