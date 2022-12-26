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
# default) or negative, `show` queries the width of the terminal and use it as
# the width, falling back to 80 if the query fails.
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
fn show {|symbol &width=0| }

# Outputs the Markdown source of the documentation for `$symbol` as a string
# value. The `$symbol` arguments follows the same format as
# [`doc:show`](#doc:show).
#
# Examples:
#
# ```elvish-transcript
# ~> doc:source put
# ▶ "... omitted "
# ```
fn source {|symbol| }
