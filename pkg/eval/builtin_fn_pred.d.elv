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

# Outputs the number -1 if `$a` is smaller than `$b`, 0 if `$a` is equal to
# `$b`, and 1 if `$a` is greater than `$b`.
#
# If `$a` and `$b` have the same type and that type is listed below, they are
# compared accordingly:
#
# -   Booleans: `$false` is smaller than `$true`.
#
# -   Typed numbers: Compared numerically, consistent with the [number
#     comparison commands](#num-cmp), except that `NaN` values are considered
#     equal to each other and smaller than all other numbers.
#
# -   Strings: Compared lexicographically by bytes, consistent with the
#     [string comparison commands](#str-cmp). For UTF-8 encoded strings, this is
#     equivalent to comparing by codepoints.
#
#     Beware that strings that look like numbers are compared as strings, not
#     numbers.
#
# -   Lists: Compared lexicographically by elements, with elements compared
#     recursively.
#
# Otherwise, if `eq $a $b` is true, `compare $a $b` outputs the number 0.
#
# For other cases, the behavior depends on the `&total` option:
#
# -   If it is `$false` (the default), `compare` throws an exception complaning
#     that the two values can't be compared.
#
# -   If it is `$true`, `compare` uses an artificial [total
#     order](https://en.wikipedia.org/wiki/Total_order) derived from the
#     following rules:
#
#     -   If they have the same type, use the rules above for comparing
#     homogeneous types.
#
#     -   If they don't have the same type, compare their types and output -1
#     or 1.
#
#         The ordering between Elvish types is unspecified, but it is guaranteed
#         to be consistent during the same Elvish session. For example, if
#         `compare &total $a $b` outputs -1 when `$a` is a number and `$b` is a
#         string, it will always output -1 for such pairs.
#
#     This artificial total order is useful when sorting values of mixed
#     types. For example, implicitly when `pprint` handles composite values
#     (e.g., lists and maps) containing mixed types.
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
# ~> compare a (num 10)
# Exception: bad value: inputs to "compare" or "order" must be comparable values, but is uncomparable values
# [tty 3]:1:1: compare a (num 10)
# ~> compare &total a (num 10)
# ▶ (num 1)
# ~> compare &total (num 10) a
# ▶ (num -1)
# ```
#
# See also [`order`]().
fn compare {|&total=$false a b| }
