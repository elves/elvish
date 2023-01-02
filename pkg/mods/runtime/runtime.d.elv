# A list containing
# [module search directories](command.html#module-search-directories).
#
# This variable is read-only.
var lib-dirs

# Path to the [RC file](command.html#rc-file), ignoring any possible overrides
# by command-line flags and available in non-interactive mode.
#
# If there was an error in determining the path of the RC file, this variable
# is `$nil`.
#
# This variable is read-only.
#
# See also [`$runtime:effective-rc-path`]().
var rc-path

# Path to the [RC path](command.html#rc-file) that is actually used for this
# Elvish session:
#
# - If Elvish is running non-interactively or invoked with the `-norc` flag,
#   this variable is `$nil`.
#
# - If Elvish is invoked with the `-rc` flag, this variable contains the
#   absolute path of the argument.
#
# - Otherwise (when Elvish is running interactively and invoked without
#   `-norc` or `-rc`), this variable has the same value as `$rc-path`.
#
# This variable is read-only.
#
# See also [`$runtime:rc-path`]().
var effective-rc-path

# The path to the Elvish binary.
#
# If there was an error in determining the path, this variable is `$nil`.
#
# This variable is read-only.
var elvish-path
