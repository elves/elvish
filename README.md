# Elvish: Expressive Programming Language + Versatile Interactive Shell

[![CI status](https://github.com/elves/elvish/workflows/CI/badge.svg)](https://github.com/elves/elvish/actions?query=workflow%3ACI)
[![FreeBSD & gccgo test status](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=CI2)](https://cirrus-ci.com/github/elves/elvish/master)
[![Test Coverage](https://img.shields.io/codecov/c/github/elves/elvish/master.svg?logo=Codecov&label=coverage)](https://app.codecov.io/gh/elves/elvish/tree/master)
[![Go Reference](https://pkg.go.dev/badge/src.elv.sh@master.svg)](https://pkg.go.dev/src.elv.sh@master)
[![Packaging status](https://repology.org/badge/tiny-repos/elvish.svg)](https://repology.org/project/elvish/versions)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/ElvishShell)

Elvish is an expressive programming language and a versatile interactive shell,
combined into one seamless package. It runs on Linux, BSDs, macOS and Windows.

Despite its pre-1.0 status, it is already suitable for most daily interactive
use.

User groups (all connected thanks to [Matrix](https://matrix.org)):
[![Gitter](https://img.shields.io/badge/gitter-elves/elvish-blue.svg?logo=gitter-white)](https://gitter.im/elves/elvish)
[![Telegram Group](https://img.shields.io/badge/telegram-@elvish-blue.svg)](https://telegram.me/elvish)
[![#elvish on libera.chat](https://img.shields.io/badge/libera.chat-%23elvish-blue.svg)](https://web.libera.chat/#elvish)
[![#users:elv.sh](https://img.shields.io/badge/matrix-%23users:elv.sh-blue.svg)](https://matrix.to/#/#users:elv.sh)

## Documentation

Documentation for Elvish lives on the official website https://elv.sh,
including:

-   [Learning material](https://elv.sh/learn)

-   [Reference docs](https://elv.sh/ref), including the
    [language reference](https://elv.sh/ref/language.html),
    [the `elvish` command](https://elv.sh/ref/command.html), and all the modules
    in the standard library

-   [Blog posts](https://elv.sh/blog), including release notes

The source for the documentation is in the
[website](https://github.com/elves/elvish/tree/master/website) directory.

## License

All source files use the BSD 2-clause license (see [LICENSE](LICENSE)), except
for the following:

-   Files in [pkg/diff](pkg/diff) and [pkg/rpc](pkg/rpc) are released under the
    BSD 3-clause license, since they are copied from
    [Go's source code](https://github.com/golang/go). See
    [pkg/diff/LICENSE](pkg/diff/LICENSE) and [pkg/rpc/LICENSE](pkg/rpc/LICENSE).

-   Files in [pkg/persistent](pkg/persistent) and its subdirectories are
    released under EPL 1.0, since they are partially derived from
    [Clojure's source code](https://github.com/clojure/clojure). See
    [pkg/persistent/LICENSE](pkg/persistent/LICENSE).

-   Files in [pkg/md/spec](pkg/md/spec) are released under the Creative Commons
    CC-BY-SA 4.0 license, since they are derived from
    [the CommonMark spec](https://github.com/commonmark/commonmark-spec). See
    [pkg/md/spec/LICENSE](pkg/md/spec/LICENSE).

## Building Elvish

Most users do not need to build Elvish from source. Prebuilt binaries for the
latest commit are provided for
[Linux amd64](https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz),
[macOS amd64](https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz),
[macOS arm64](https://dl.elv.sh/darwin-arm64/elvish-HEAD.tar.gz),
[Windows amd64](https://dl.elv.sh/windows-amd64/elvish-HEAD.zip) and
[many other platforms](https://elv.sh/get).

To build Elvish from source, you need

-   A supported OS: Linux, {Free,Net,Open}BSD, macOS, or Windows 10. Windows 10
    support is experimental.

-   Go >= 1.20.

To build Elvish from source, run one of the following commands:

```sh
go install src.elv.sh/cmd/elvish@master # Install latest commit
go install src.elv.sh/cmd/elvish@latest # Install latest released version
go install src.elv.sh/cmd/elvish@v0.18.0 # Install a specific version
```

### Controlling the installation location

The
[`go install`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
command installs Elvish to `$GOBIN`; the binary name is `elvish`. You can
control the installation location by overriding `$GOBIN`, for example by
prepending `env GOBIN=...` to the `go install` command.

If `$GOBIN` is not set, the installation location defaults to `$GOPATH/bin`,
which in turn defaults to `~/go/bin` if `$GOPATH` is also not set.

The installation directory is probably not in your OS's default `$PATH`. You
should either either add it to `$PATH`, or manually copy the Elvish binary to a
directory already in `$PATH`.

### Building a variant

Elvish has several *build variants* with slightly different feature sets. For
example, the `withpprof` build variant has
[profiling support](https://pkg.go.dev/runtime/pprof).

These build variants are just alternative main packages. For example, to build
the `withpprof` variant, run the following command (change the part after `@` to
get different versions):

```sh
go install src.elv.sh/cmd/withpprof/elvish@master
```

### Building from a local source tree

If you are modifying Elvish's source code, you will want to clone Elvish's Git
repository and build Elvish from the local source tree instead. To do this, run
the following from the root of the source tree:

```sh
go install ./cmd/elvish
```

There is no need to specify a version like `@master`; when inside a source tree,
`go install` will always use the whatever source code is present.

See [CONTRIBUTING.md](CONTRIBUTING.md) for more notes for contributors.

### Building with experimental plugin support

Elvish has experimental support for building and importing plugins, modules
written in Go. It relies on Go's [plugin support](https://pkg.go.dev/plugin),
which is only available on a few platforms.

Plugin support requires building Elvish with [cgo](https://pkg.go.dev/cmd/cgo).
The official [prebuilt binaries](https://elv.sh/get) are built without cgo for
compatibility and reproducibility, but by default the Go toolchain builds with
cgo enabled.

If you have built Elvish from source on a platform with plugin support, your
Elvish build probably already supports plugins. To force cgo to be used when
building Elvish, you can do the following:

```sh
env CGO_ENABLED=1 go install ./cmd/elvish
```

To build a plugin, see this [example](https://github.com/elves/sample-plugin).

## Packaging Elvish

See [PACKAGING.md](PACKAGING.md) for notes for packagers.

## Contributing to Elvish

See [CONTRIBUTING.md](CONTRIBUTING.md) for notes for contributors.

## Reporting security issues

See [SECURITY.md](SECURITY.md) for how to report security issues.
