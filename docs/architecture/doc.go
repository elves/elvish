/*
# Overview

This file documents how Elvish's codebase is structured on a high level. You
can read it either in a code editor, or in a godoc viewer such as
https://pkg.go.dev/src.elv.sh@master/docs/architecture.

Elvish is a Go project. If you are not familiar with how Go code is
organized, start with [how to write Go code].

Go code in the Elvish repo lives under two directories:

  - The cmd directory contains Elvish's entrypoints, but it contains very little code.
  - The pkg directory has most of Elvish's Go code. It has a lot of
    subdirectories, so it can be a bit hard to find your bearing just by
    exploring the file tree.

We will cover the cmd directory first, and then focus on the most important
subdirectories under pkg.

The Elvish repo also contains other directories. They are not technically
part of the Go program, so we won't cover them here. Read their respective
README files to learn more.

# Module, package and symbol names

Elvish's module name is [src.elv.sh]. You can think of it as an alias to where
the code is actually hosted (currently [github.com/elves/elvish]).

The import paths of all the packages start with the module name [src.elv.sh].
For example, the import path of the package in pkg/parse is
[src.elv.sh/pkg/parse].

When referring to a symbol from a package, we'll use just the last component of
the package's import path. For example, the Evaler type from the
[src.elv.sh/pkg/eval] package is simply [eval.Evaler]. (This is consistent with
Go's syntax.)

# Entrypoints (cmd/elvish and pkg/prog)

The default entrypoint of Elvish is [src.elv.sh/cmd/elvish]. It has a main
function that does the following:

  - Assemble a "composite program" from multiple subprograms, most
    notably [shell.Program].
  - Call [prog.Run].

You can read about the advantage of this approach in the godoc of
[src.elv.sh/pkg/prog].

There are other main packages, like [src.elv.sh/cmd/withpprof/elvish]. They
follow the same structure and only differ in which subprograms they include.

# The shell subprogram (pkg/shell)

The shell subprogram has two slightly different "modes", interactive and
non-interactive, depending on the command-line arguments. The doc for
[shell.Program] contains more details.

In both modes, the shell subprogram uses the interpreter implemented in
[src.elv.sh/pkg/eval] to evaluate code.

In interactive mode, the shell also uses the line editor implemented in
[src.elv.sh/pkg/edit] to read commands interactively. Some features of the
editor depend on persistent storage; the shell subprogram also takes care of
initializing that, using [src.elv.sh/pkg/daemon].

# The interpreter (pkg/eval)

The [src.elv.sh/pkg/eval] package is perhaps the most important package in
Elvish, as it implements the Elvish language and the builtin module.

The interpreter is represented by [eval.Evaler], which is created with
[eval.NewEvaler]. The method [eval.Evaler.Eval] (yes, that's 3 "evals"s)
evaluates Elvish code, and does so in several steps:

 1. Invoke the parser to get an AST.
 2. Compile the AST into an "operation tree".
 3. Run the operation tree.

This approach is chosen mainly for its simplicity. It's probably not very
performant.

The compilation of each AST node into its corresponding operation node, as well
as how each operation node runs, is defined in the several compile_*.go files.
These files are where most of the language semantics is implemented.

Another sizable chunk of this package is the various builtin_fn_*.go files,
which implement functions of the builtin module. These may be moved to a
different package in future.

Some other packages important for the interpreter are:

  - [src.elv.sh/pkg/eval/vals] implements a standard set of operations for
    Elvish values.
  - [src.elv.sh/pkg/persistent] implements Elvish's lists and maps, modeled
    after [Clojure's vectors and maps].
  - Subdirectories of [src.elv.sh/pkg/mods] implement the various builtin
    modules.

# The parser (pkg/parse)

The [src.elv.sh/pkg/parse] package implements parsing of Elvish code, with the
[parse.Parse] as the entrypoint.

The parsing algorithm is a handwritten [recursive descent] one, with the
slightly unusual property that there's no separate tokenization phase. Read the
package's godoc for more details.

# The editor (pkg/edit)

The [src.elv.sh/pkg/edit] package contains Elvish's interactive line editor,
represented by [edit.Editor]. The traditional term "line editor" is a bit of a
misnomer; modern line editors (including Elvish's) are similar to full-blown TUI
applications like Vim, except that they usually restrict themselves to the last
N lines of the terminal rather than the entire screen.

The editor is built on top of the more low-level [src.elv.sh/pkg/cli] package
(which is also a bit of a misnomer), in particular the [cli.App] type.

The entire TUI stack is due for a rewrite soon.

The editor relies on persistent storage for features like the directory history
and the command history. As mentioned above, the initialization of the storage
is done in pkg/shell, using pkg/daemon.

# The storage daemon (pkg/daemon)

Support for persistent storage is is currently provided by a storage daemon. The
[src.elv.sh/pkg/daemon] packages implements two things:

  - A subprogram implementing the storage daemon ([daemon.Program]).
  - A client to talk to the daemon (returned by [daemon.Activate]).

The daemon is launched and terminated on demand:

  - The first interactive Elvish shell launches the daemon.
  - Subsequent interactive shells talk to the same daemon.
  - When the last interactive Elvish shell quits, the daemon also quits.

Internally, the daemon uses [bbolt] as the database engine.

In future (subject to evaluation) Elvish might get a custom database, and the
daemon might go away.

# Closing remarks

This should have given you a rough idea of the most important bits of Elvish's
implementation. The implementation prioritizes readability, and most exported
symbols are documented, so feel free to dive into the source code!

If you have questions, feel free to ask in the user group or DM xiaq.

[how to write Go code]: https://go.dev/doc/code
[github.com/elves/elvish]: https://github.com/elves/elvish
[src.elv.sh]: https://src.elv.sh
[Clojure's vectors and maps]: https://clojure.org/reference/data_structures
[recursive descent]: https://en.wikipedia.org/wiki/Recursive_descent_parser
[bbolt]: https://github.com/etcd-io/bbolt
*/
package architecture

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
)

var (
	_ = new(shell.Program)
	_ = prog.Run

	_ = eval.NewEvaler
	_ = (*eval.Evaler).Eval

	_ = parse.Parse

	_ = new(edit.Editor)
	_ = new(cli.App)

	_ = new(daemon.Program)
	_ = daemon.Activate
)
