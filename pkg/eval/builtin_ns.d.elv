# A blackhole variable.
#
# Values assigned to it will be discarded. Referencing it always results in $nil.
var _

#//skip-test
# A list containing command-line arguments. Analogous to `argv` in some other
# languages. Examples:
#
# ```elvish-transcript
# ~> echo 'put $args' > args.elv
# ~> elvish args.elv foo -bar
# ▶ [foo -bar]
# ~> elvish -c 'put $args' foo -bar
# ▶ [foo -bar]
# ```
#
# As demonstrated above, this variable does not contain the name of the script
# used to invoke it. For that information, use the `src` command.
#
# See also [`src`]().
var args

# The boolean false value.
var false

# The special value used by `?()` to signal absence of exceptions.
var ok

# A special value useful for representing the lack of values.
var nil

# A list of search paths, kept in sync with `$E:PATH`. It is easier to use than
# `$E:PATH`.
var paths

# The process ID of the current Elvish process.
var pid

# The present working directory. Setting this variable has the same effect as
# `cd`. This variable is most useful in a temporary assignment.
#
# Example:
#
# ```elvish
# ## Updates all git repositories
# use path
# for x [*/] {
#   tmp pwd = $x
#   if (path:is-dir .git) {
#     git pull
#   }
# }
# ```
#
# Etymology: the `pwd` command.
#
# See also [`cd`]().
var pwd

# The boolean true value.
var true

# A map that exposes information about the Elvish binary. Running `put
# $buildinfo | to-json` will produce the same output as `elvish -buildinfo
# -json`.
#
# See also [`$version`]().
var buildinfo

# The full version of the Elvish binary as a string. This is the same information reported by
# `elvish -version` and the value of `$buildinfo[version]`.
#
# **Note:** In general it is better to perform functionality tests rather than testing `$version`.
# For example, do something like
#
# ```elvish
# has-key $builtin: new-var
# ```
#
# to test if variable `new-var` is available rather than comparing against `$version` to see if the
# elvish version is equal to or newer than the version that introduced `new-var`.
#
# See also [`$buildinfo`]().
var version
