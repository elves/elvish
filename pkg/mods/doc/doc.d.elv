# Shows documentation for `$symbol` in the terminal.
#
# If `$symbol` starts with `$`, it is treated as a variable. Otherwise it is
# treated as a function.
#
# Symbols in a module should be specified using a qualified name as if the
# module is imported without renaming, like `doc:source`. Symbols in the builtin
# module can be specified either in the unqualified form (like `put`) or with
# the explicit `builtin:` namespace (like `builtin:put`).
#
# The `&width` option specifies the width to wrap the output to. If it is 0 (the
# default) or negative, `show` queries the terminal width of the standard output
# and use it as the width, falling back to 80 if the query fails (for example
# when the standard output is not a terminal).
#
# This command is roughly equivalent to `md:show &width=$width (doc:show
# $symbol)`, but has some extra processing of relative links to point them to
# the Elvish website.
#
# Examples:
#
# ```elvish-transcript
# ~> doc:show put
# [ omitted ]
# ~> doc:show builtin:put
# [ omitted ]
# ~> doc:show '$paths'
# [ omitted ]
# ~> doc:show doc:show
# [ omitted ]
# ```
#
# See also [`md:show`]().
fn show {|symbol &width=0| }

# Finds symbols whose documentation contains all strings in `$queries`.
#
# The search is done on a version of the documentation with no markup and soft
# line breaks converted to spaces. For example, if the Markdown source of a
# symbol contains `foo *bar* [link](dest.html)`, with possible soft line breaks
# in between, it will be matched by a query of `foo bar link`.
#
# The output shows each symbol that matches, followed by an excerpt of their
# documentation with the matched queries highlighted.
#
# Examples:
#
# ```elvish-transcript
# ~> doc:find namespace
# ns:
#   Constructs a namespace from $map, using the keys as variable names and the values as their values. …
# has-env:
#   … This command has no equivalent operation using the E: namespace (but see https://b.elv.sh/1026).
# eval:
#   … The evaluation happens in a new, restricted namespace, whose initial set of variables can be specified by the &ns option. …
# [ … more output omitted … ]
# ~> doc:find namespace REPL
# edit:add-var:
#   Defines a new variable in the interactive REPL with an initial value. …
#   This is most useful for modules to modify the REPL namespace. …
# ```
fn find {|@queries| }

# Outputs the Markdown source of the documentation for `$symbol` as a string
# value. The `$symbol` arguments follows the same format as
# [`doc:show`]().
#
# Examples:
#
# ```elvish-transcript
# ~> doc:source put
# ▶ "... omitted "
# ```
fn source {|symbol| }
