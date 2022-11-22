# Sets an environment variable to the given value. Calling `set-env VAR_NAME
# value` is similar to `set E:VAR_NAME = value`, but allows the variable name
# to be dynamic.
#
# Example:
#
# ```elvish-transcript
# ~> set-env X foobar
# ~> put $E:X
# ▶ foobar
# ```
#
# @cf get-env has-env unset-env
fn set-env {|name value| }

# Unset an environment variable. Calling `unset-env VAR_NAME` is similar to
# `del E:VAR_NAME`, but allows the variable name to be dynamic.
#
# Example:
#
# ```elvish-transcript
# ~> set E:X = foo
# ~> unset-env X
# ~> has-env X
# ▶ $false
# ~> put $E:X
# ▶ ''
# ```
#
# @cf has-env get-env set-env
fn unset-env {|name| }

# Test whether an environment variable exists. This command has no equivalent
# operation using the `E:` namespace (but see https://b.elv.sh/1026).
#
# Examples:
#
# ```elvish-transcript
# ~> has-env PATH
# ▶ $true
# ~> has-env NO_SUCH_ENV
# ▶ $false
# ```
#
# @cf get-env set-env unset-env
fn has-env {|name| }

# Gets the value of an environment variable. Throws an exception if the
# environment variable does not exist.
#
# Calling `get-env VAR_NAME` is similar to `put $E:VAR_NAME`, but allows the
# variable name to be dynamic, and throws an exception instead of producing an
# empty string for nonexistent environment variables.
#
# Examples:
#
# ```elvish-transcript
# ~> get-env LANG
# ▶ zh_CN.UTF-8
# ~> get-env NO_SUCH_ENV
# Exception: non-existent environment variable
# [tty], line 1: get-env NO_SUCH_ENV
# ```
#
# @cf has-env set-env unset-env
fn get-env {|name| }
