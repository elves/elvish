# Elvish: Expressive Programming Language + Versatile Interactive Shell

[![CI status](https://github.com/elves/elvish/workflows/CI/badge.svg)](https://github.com/elves/elvish/actions?query=workflow%3ACI)
[![FreeBSD & gccgo test status](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=CI2)](https://cirrus-ci.com/github/elves/elvish/master)
[![Test Coverage](https://img.shields.io/codecov/c/github/elves/elvish/master.svg?logo=Codecov&label=coverage)](https://app.codecov.io/gh/elves/elvish/branch/master)
[![Go Reference](https://pkg.go.dev/badge/src.elv.sh@master.svg)](https://pkg.go.dev/src.elv.sh@master)
[![Packaging status](https://repology.org/badge/tiny-repos/elvish.svg)](https://repology.org/project/elvish/versions)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/ElvishShell)

Elvish is an expressive programming language and a versatile interactive shell,
combined into one seamless package. It runs on Linux, BSDs, macOS and Windows.

Despite its pre-1.0 status, it is already suitable for most daily interactive
use.

**Visit the official website https://elv.sh for prebuilt binaries, blog posts,
documentation and other resources.**

User groups (all connected thanks to [Matrix](https://matrix.org)):
[![Gitter](https://img.shields.io/badge/gitter-elves/elvish-blue.svg?logo=gitter-white)](https://gitter.im/elves/elvish)
[![Telegram Group](https://img.shields.io/badge/telegram-@elvish-blue.svg)](https://telegram.me/elvish)
[![#elvish on libera.chat](https://img.shields.io/badge/libera.chat-%23elvish-blue.svg)](https://web.libera.chat/#elvish)
[![#users:elves.sh](https://img.shields.io/badge/matrix-%23users:elv.sh-blue.svg)](https://matrix.to/#/#users:elves.sh)

## Building Elvish

Most users do not need to build Elvish from source. Prebuilt binaries for the
latest commit are provided for
[Linux amd64](https://dl.elv.sh/linux-amd64/elvish-HEAD.tar.gz),
[macOS amd64](https://dl.elv.sh/darwin-amd64/elvish-HEAD.tar.gz),
[Windows amd64](https://dl.elv.sh/windows-amd64/elvish-HEAD.zip) and
[many other platforms](https://elv.sh/get).

To build Elvish from source, you need

-   A supported OS: Linux, {Free,Net,Open}BSD, macOS, or Windows 10.

    **NOTE**: Windows 10 support is experimental.

-   Go >= 1.16.

To build Elvish from source, follow these steps:

```sh
# 1. Start from any directory you want to store Elvish's source code
# 2. Clone the Git repository
git clone https://github.com/elves/elvish
# 3. Change into the repository
cd elvish
# 4. Build and install Elvish
make get
```

This will install Elvish to `~/go/bin` (or `$GOPATH/bin` if you have set
`$GOPATH`). You might want to add the directory to your `PATH`.

To install it elsewhere, override `ELVISH_MAKE_BIN` in the `make` command:

```sh
make get ELVISH_MAKE_BIN=./elvish # Install to the repo root
make get ELVISH_MAKE_BIN=/usr/local/bin/elvish # Install to /usr/local/bin
```

### Experimental plugin support

Elvish has experimental support for building and importing plugins, modules
written in Go.

However, since plugin support relies on dynamic linking, it is not enabled in
the official prebuilt binaries. You need to build Elvish from source, with
`ELVISH_PLUGIN_SUPPORT=1`:

```sh
make get ELVISH_PLUGIN_SUPPORT=1
```

To build a plugin, see this [example](https://github.com/elves/sample-plugin).

## Packaging Elvish

See [PACKAGING.md](PACKAGING.md) for notes for packagers.

## Contributing to Elvish

See [CONTRIBUTING.md](CONTRIBUTING.md) for notes for contributors.
