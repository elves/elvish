# Convert a value to boolean. In Elvish, only `$false`, `$nil`, and errors are
# booleanly false. Everything else, including 0, empty strings and empty lists,
# is booleanly true:
#
# ```elvish-transcript
# ~> bool $true
# ▶ $true
# ~> bool $false
# ▶ $false
# ~> bool $ok
# ▶ $true
# ~> bool ?(fail haha)
# ▶ $false
# ~> bool ''
# ▶ $true
# ~> bool []
# ▶ $true
# ~> bool abc
# ▶ $true
# ```
#
# See also [`not`]().
fn bool {|value| }

# Boolean negation. Examples:
#
# ```elvish-transcript
# ~> not $true
# ▶ $false
# ~> not $false
# ▶ $true
# ~> not $ok
# ▶ $false
# ~> not ?(fail error)
# ▶ $true
# ```
#
# **Note**: The related logical commands `and` and `or` are implemented as
# [special commands](language.html#special-commands) instead, since they do not
# always evaluate all their arguments. The `not` command always evaluates its
# only argument, and is thus a normal command.
#
# See also [`bool`]().
fn not {|value| }

# Determine whether all `$value`s have the same identity. Writes `$true` when
# given no or one argument.
#
# The definition of identity is subject to change. Do not rely on its behavior.
#
# ```elvish-transcript
# ~> is a a
# ▶ $true
# ~> is a b
# ▶ $false
# ~> is [] []
# ▶ $true
# ~> is [a] [a]
# ▶ $false
# ```
#
# See also [`eq`]().
#
# Etymology: [Python](https://docs.python.org/3/reference/expressions.html#is).
fn is {|@values| }

# Determines whether all `$value`s are equal. Writes `$true` when
# given no or one argument.
#
# Two values are equal when they have the same type and value.
#
# For complex data structures like lists and maps, comparison is done
# recursively. A pseudo-map is equal to another pseudo-map with the same
# internal type (which is not exposed to Elvish code now) and value.
#
# ```elvish-transcript
# ~> eq a a
# ▶ $true
# ~> eq [a] [a]
# ▶ $true
# ~> eq [&k=v] [&k=v]
# ▶ $true
# ~> eq a [b]
# ▶ $false
# ```
#
# See also [`is`]() and [`not-eq`]().
#
# Etymology: [Perl](https://perldoc.perl.org/perlop.html#Equality-Operators).
fn eq {|@values| }

# Determines whether `$a` and `$b` are not equal. Equivalent to `not (eq $a $b)`.
#
# ```elvish-transcript
# ~> not-eq 1 2
# ▶ $true
# ~> not-eq 1 1
# ▶ $false
# ```
#
# See also [`eq`]().
fn not-eq {|a b| }

# Outputs the number -1 if `$a` is smaller than `$b`, 0 if `$a` is equal to
# `$b`, and 1 if `$a` is greater than `$b`.
#
# The following algorithm is used:
#
# 1.  If `$a` and `$b` have the same type and that type is listed below, they are
#     compared accordingly:
#
#      -   Booleans: `$false` is smaller than `$true`.
#
#      -   Typed numbers: Compared numerically, consistent with [`<`](#num-lt)
#          and [`<=`](#num-le), except that `NaN` values are considered equal to
#          each other and smaller than all other numbers.
#
#      -   Strings: Compared lexicographically by bytes, consistent with
#          [`<s`](#str-lt) and [`<=s`](#str-le). For UTF-8 encoded strings, this
#          is equivalent to comparing by codepoints.
#
#          Beware that strings that look like numbers are compared as strings,
#          not numbers.
#
#      -   Lists: Compared lexicographically by elements, with elements compared
#          recursively.
#
# 2.  If `eq $a $b` is true, `compare $a $b` outputs the number 0.
#
# 3.  Otherwise the behavior depends on the `&total` option:
#
#     -   If it is `$false` (the default), `compare` throws an exception
#         complaning that the two values can't be compared.
#
#     -   If it is `$true`, `compare` compares the _types_ of `$a` and `$b`: if
#         they have the same type, it outputs 0; if they have different types,
#         it outputs -1 and 1 depending on which type comes first in an internal
#         ordering of all types.
#
#         The internal ordering of all types is unspecified, but it is
#         guaranteed to be consistent during the same Elvish session. For
#         example, if `compare &total $a $b` outputs -1 when `$a` is a number
#         and `$b` is a string, it will always output -1 for such pairs.
#
#         This creates an artificial [total
#         order](https://en.wikipedia.org/wiki/Total_order), which is mainly
#         useful for sorting values of mixed types.
#
# Examples comparing values of the same type:
#
# ```elvish-transcript
# ~> compare a b
# ▶ (num -1)
# ~> compare b a
# ▶ (num 1)
# ~> compare x x
# ▶ (num 0)
# ~> compare (num 10) (num 1)
# ▶ (num 1)
# ```
#
# Examples comparing values of different types:
#
# ```elvish-transcript
# //skip-test
# ~> compare a (num 10)
# Exception: bad value: inputs to "compare" or "order" must be comparable values, but is uncomparable values
#   [tty]:1:1-18: compare a (num 10)
# ~> compare &total a (num 10)
# ▶ (num -1)
# ~> compare &total (num 10) a
# ▶ (num 1)
# ```
#
# See also [`order`]().
fn compare {|&total=$false a b| }
