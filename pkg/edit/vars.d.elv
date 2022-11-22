# Defines a new variable in the interactive REPL with an initial value. The new variable becomes
# available during the next REPL cycle.
#
# Equivalent to running `var $name = $init` at a REPL prompt, but `$name` can be
# dynamic.
#
# This is most useful for modules to modify the REPL namespace. Example:
#
# ```elvish-transcript
# ~> cat .config/elvish/lib/a.elv
# for i [(range 10)] {
#   edit:add-var foo$i $i
# }
# ~> use a
# ~> put $foo1 $foo2
# ▶ (num 1)
# ▶ (num 2)
# ```
#
# Note that if you use a variable as the `$init` argument, `edit:add-var`
# doesn't add the variable "itself" to the REPL namespace. The variable in the
# REPL namespace will have the initial value set to the variable's value, but
# it is not an alias of the original variable:
#
# ```elvish-transcript
# ~> cat .config/elvish/lib/b.elv
# var foo = foo
# edit:add-var foo $foo
# ~> use b
# ~> put $foo
# ▶ foo
# ~> set foo = bar
# ~> echo $b:foo
# foo
# ```
#
# ### Importing definition from a module into the REPL
#
# One common use of this command is to put the definitions of functions intended for REPL use in a
# module instead of your [`rc.elv`](command.html#rc-file). For example, if you want to define `ll`
# as `ls -l`, you can do so in your `rc.elv` directly:
#
# ```elvish
# fn ll {|@a| ls -l $@a }
# ```
#
# But if you move the definition into a module (say `util.elv` in one of the
# [module search directories](command.html#module-search-directories), this
# function can only be used as `util:ll` (after `use util`). To make it usable
# directly as `ll`, you can add the following to `util.elv`:
#
# ```elvish
# edit:add-var ll~ $ll~
# ```
#
# ### Conditionally importing a module
#
# Another use case is to add a module or function to the REPL namespace
# conditionally. For example, to only import [the `unix` module](unix.html)
# when actually running on UNIX, a straightforward solution is to do the
# following in `rc.elv`:
#
# ```elvish
# use platform
# if $platform:is-unix {
#   use unix
# }
# ```
#
# This doesn't work however, since what `use` does is introducing a variable
# named `$unix:`. Since all variables in Elvish are lexically scoped, the
# `$unix:` variable is only valid inside the `if` block.
#
# This can be fixed by explicitly introducing the `$unix:` variable to the REPL
# namespace. The following works both from `rc.elv` and from a module:
#
# ```elvish
# use platform
# if $platform:is-unix {
#   use unix
#   edit:add-var unix: $unix:
# }
# ```
fn add-var {|name init| }

# Takes a map from strings to arbitrary values. Equivalent to calling
# `edit:add-var` for each key-value pair in the map.
fn add-vars {|map| }
