#elvdoc:var lib-dirs
#
# A list containing
# [module search directories](command.html#module-search-directories).
#
# This variable is read-only.

#elvdoc:var rc-path
#
# Path to the [RC file](command.html#rc-file), ignoring any possible overrides
# by command-line flags and available in non-interactive mode.
#
# If there was an error in determining the path of the RC file, this variable
# is `$nil`.
#
# This variable is read-only.
#
# @cf $runtime:effective-rc-path

#elvdoc:var effective-rc-path
#
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
# @cf $runtime:rc-path

#elvdoc:var elvish-path
#
# The path to the Elvish binary.
#
# If there was an error in determining the path, this variable is `$nil`.
#
# This variable is read-only.
