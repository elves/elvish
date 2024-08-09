# Elvish

[![CI status](https://github.com/elves/elvish/workflows/CI/badge.svg)](https://github.com/elves/elvish/actions?query=workflow%3ACI)
[![FreeBSD & gccgo test status](https://img.shields.io/cirrus/github/elves/elvish?logo=Cirrus%20CI&label=CI2)](https://cirrus-ci.com/github/elves/elvish/master)
[![Test Coverage](https://img.shields.io/codecov/c/github/elves/elvish/master.svg?logo=Codecov&label=coverage)](https://app.codecov.io/gh/elves/elvish/tree/master)
[![Go Reference](https://pkg.go.dev/badge/src.elv.sh@master.svg)](https://pkg.go.dev/src.elv.sh@master)
[![Packaging status](https://repology.org/badge/tiny-repos/elvish.svg)](https://repology.org/project/elvish/versions)

[![Forum](https://img.shields.io/badge/forum-bbs.elv.sh-5b5.svg?logo=discourse)](https://bbs.elv.sh)
[![Twitter](https://img.shields.io/badge/twitter-@ElvishShell-blue.svg?logo=x)](https://twitter.com/ElvishShell)

[![Telegram Group](https://img.shields.io/badge/telegram-Elvish-blue.svg?logo=telegram&logoColor=white)](https://t.me/+Pv5ZYgTXD-YaKwcP)
[![Discord server](https://img.shields.io/badge/discord-Elvish-blue.svg?logo=discord&logoColor=white)](https://discord.gg/jrmuzRBU8D)
[![#users:elv.sh](https://img.shields.io/badge/matrix-%23users:elv.sh-blue.svg?logo=matrix)](https://matrix.to/#/#users:elv.sh)
[![#elvish on libera.chat](https://img.shields.io/badge/libera.chat-%23elvish-blue.svg?logo=liberadotchat&logoColor=white)](https://web.libera.chat/#elvish)
[![Gitter](https://img.shields.io/badge/gitter-elves%2Felvish-blue.svg?logo=gitter)](https://gitter.im/elves/elvish)

(Chat rooms are all bridged together thanks to [Matrix](https://matrix.org).)

Elvish is:

-   A powerful scripting language.

-   A shell with useful interactive features built-in.

-   A statically linked binary for Linux, BSDs, macOS or Windows.

Elvish is pre-1.0. This means that breaking changes will still happen from time
to time, but it's stable enough for both scripting and interactive use.

## Documentation

[![User docs](https://img.shields.io/badge/User_Docs-37a779?style=for-the-badge)](https://elv.sh)

User docs are hosted on Elvish's website, [elv.sh](https://elv.sh). This
includes [how to install Elvish](https://elv.sh/get/),
[tutorials](https://elv.sh/learn/), [reference pages](https://elv.sh/ref/), and
[news](https://elv.sh/blog/).

[![Development docs](https://img.shields.io/badge/Development_Docs-blue?style=for-the-badge)](./docs)

Development docs are in [./docs](./docs).

[![Awesome Elvish](https://img.shields.io/badge/Awesome_Elvish-orange?style=for-the-badge)](https://github.com/elves/awesome-elvish)

Awesome Elvish packages and tools that support Elvish.

## License

All source files use the BSD 2-clause license (see [LICENSE](LICENSE)), except
for the following:

-   Files in [pkg/diff](pkg/diff) and [pkg/rpc](pkg/rpc) are released under the
    BSD 3-clause license, since they are derived from
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
