# Building Elvish from source

To build Elvish from source, you need

-   A supported OS: Linux, {Free,Net,Open}BSD, macOS, or Windows 10. Windows 10
    support is experimental.

-   Go >= 1.21.0.

To build Elvish from source, run one of the following commands:

```sh
go install src.elv.sh/cmd/elvish@master # Install latest commit
go install src.elv.sh/cmd/elvish@latest # Install latest released version
go install src.elv.sh/cmd/elvish@v0.18.0 # Install a specific version
```

## Controlling the installation location

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

## Building an alternative entrypoint

In additional to `src.elv.sh/cmd/elvish` (which corresponds to the
[`cmd/elvish`](./cmd/elvish) directory in the repo), there are a few alternative
entrypoints, all named liked `cmd/*/elvish`, with slightly different feature
sets. (From the perspective of Go, these are just different `main` packages.)

For example, install the `cmd/withpprof/elvish` entrypoint to get
[profiling support](https://pkg.go.dev/runtime/pprof) (change the part after `@`
to get different versions):

```sh
go install src.elv.sh/cmd/withpprof/elvish@master
```

## Building from a local source tree

If you are modifying Elvish's source code, you will want to clone Elvish's Git
repository and build Elvish from the local source tree instead. To do this, run
the following from the root of the source tree:

```sh
go install ./cmd/elvish
```

There is no need to specify a version like `@master`; when inside a source tree,
`go install` will always use the whatever source code is present.

See [CONTRIBUTING.md](CONTRIBUTING.md) for more notes for contributors.

## Building with experimental plugin support

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
