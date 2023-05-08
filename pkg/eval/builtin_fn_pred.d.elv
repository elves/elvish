# Convert a value to boolean. In Elvish, only `$false` and errors are booleanly
# false. Everything else, including 0, empty strings and empty lists, is booleanly
# true:
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

# Determines whether every adjacent pair of `$value`s are not equal. Note that
# this does not imply that `$value`s are all distinct. Examples:
#
# ```elvish-transcript
# ~> not-eq 1 2 3
# ▶ $true
# ~> not-eq 1 2 1
# ▶ $true
# ~> not-eq 1 1 2
# ▶ $false
# ```
#
# See also [`eq`]().
fn not-eq {|@values| }

# Outputs -1 if `$a` < `$b`, 0 if `$a` = `$b`, and 1 if `$a` > `$b`.
#
# The following comparison algorithm is used:
#
# - Typed numbers are compared numerically. The comparison is consistent with
#   the [number comparison commands](#num-cmp), except that `NaN` values are
#   considered equal to each other and smaller than all other numbers.
#
# - Strings are compared lexicographically by bytes, consistent with the
#   [string comparison commands](#str-cmp). For UTF-8 encoded strings, this is
#   equivalent to comparing by codepoints.
#
# - Lists are compared lexicographically by elements, if the elements at the
#   same positions are comparable.
#
# If `&types` is `$false` (the default), and the ordering between two elements
# is not defined by the conditions above, i.e. if the value of `$a` or `$b` is
# not covered by any of the cases above or if they belong to different cases,
# a "bad value" exception is thrown.
#
# If `&types` is `$true` and the values are not comparable then the type of
# each value is compared. The value types have the following relationship:
# bool < num < string < list < map.
#
# Examples:
#
# ```elvish-transcript
# ~> compare a b
# ▶ (num 1)
# ~> compare b a
# ▶ (num -1)
# ~> compare x x
# ▶ (num 0)
# ~> compare (num 10) (num 1)
# ▶ (num 1)
# ~> compare (num 1) $true
# Exception: bad value: inputs to "compare" or "order" must be comparable values, but is uncomparable values
# ~> compare &types (num 1) $true
# ▶ (num 1)
# ~> compare (num 1) string
# Exception: bad value: inputs to "compare" or "order" must be comparable values, but is uncomparable values
# ~> compare &types (num 1) string
# ▶ (num -1)
# ```
#
# Beware that strings that look like numbers are treated as strings, not
# numbers. Also, unlike the other comparison commands (e.g., [`==`](#num-cmp)
# and [`<`](#num-cmp)) this command requires exactly two values. If not given
# two values an "arity mismatch" exception is thrown.
#
# See also [`order`]().
fn compare {|&types=$false a b| }
