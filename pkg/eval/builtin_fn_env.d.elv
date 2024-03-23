# Sets an environment variable to the given value. Calling `set-env VAR_NAME
# value` is similar to `set E:VAR_NAME = value`, but allows the variable name
# to be dynamic.
#
# Example:
#
# ```elvish-transcript
# //unset-env X
# ~> set-env X foobar
# ~> put $E:X
# ▶ foobar
# ```
#
# See also [`get-env`](), [`has-env`](), and [`unset-env`]().
fn set-env {|name value| }

# Unset an environment variable. Calling `unset-env VAR_NAME` is similar to
# `del E:VAR_NAME`, but allows the variable name to be dynamic.
#
# Example:
#
# ```elvish-transcript
# //unset-env X
# ~> set E:X = foo
# ~> unset-env X
# ~> has-env X
# ▶ $false
# ~> put $E:X
# ▶ ''
# ```
#
# See also [`has-env`](), [`get-env`](), and [`set-env`]().
fn unset-env {|name| }

# Test whether an environment variable exists. This command has no equivalent
# operation using the `E:` namespace (but see https://b.elv.sh/1026).
#
# Examples:
#
# ```elvish-transcript
# //set-env PATH /bin
# //unset-env NO_SUCH_ENV
# ~> has-env PATH
# ▶ $true
# ~> has-env NO_SUCH_ENV
# ▶ $false
# ```
#
# See also [`get-env`](), [`set-env`](), and [`unset-env`]().
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
# //set-env LANG zh_CN.UTF-8
# //unset-env NO_SUCH_ENV
# ~> get-env LANG
# ▶ zh_CN.UTF-8
# ~> get-env NO_SUCH_ENV
# Exception: non-existent environment variable
#   [tty]:1:1-19: get-env NO_SUCH_ENV
# ```
#
# See also [`has-env`](), [`set-env`](), and [`unset-env`]().
fn get-env {|name| }
