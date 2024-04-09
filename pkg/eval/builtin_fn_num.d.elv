#//skip-test
# Output a pseudo-random number in the interval [0, 1). Example:
#
# ```elvish-transcript
# ~> rand
# ▶ 0.17843564133528436
# ```
fn rand { }

# Constructs a [typed number](language.html#number).
#
# If the argument is a string, this command outputs the typed number the
# argument represents, or raises an exception if the argument is not a valid
# representation of a number. If the argument is already a typed number, this
# command outputs it as is.
#
# This command is usually not needed for working with numbers; see the
# discussion of [numeric commands](#numeric-commands).
#
# Examples:
#
# ```elvish-transcript
# ~> num 10
# ▶ (num 10)
# ~> num 0x10
# ▶ (num 16)
# ~> num 1/12
# ▶ (num 1/12)
# ~> num 3.14
# ▶ (num 3.14)
# ~> num (num 10)
# ▶ (num 10)
# ```
#
# See also [`exact-num`]() and [`inexact-num`]().
fn num {|string-or-number| }

# Coerces the argument to an exact number. If the argument is infinity or NaN,
# an exception is thrown.
#
# If the argument is a string, it is converted to a typed number first. If the
# argument is already an exact number, it is returned as is.
#
# Examples:
#
# ```elvish-transcript
# ~> exact-num (num 0.125)
# ▶ (num 1/8)
# ~> exact-num 0.125
# ▶ (num 1/8)
# ~> exact-num (num 1)
# ▶ (num 1)
# ```
#
# Beware that seemingly simple fractions that can't be represented precisely in
# binary can result in the denominator being a very large power of 2:
#
# ```elvish-transcript
# ~> exact-num 0.1
# ▶ (num 3602879701896397/36028797018963968)
# ```
#
# See also [`num`]() and [`inexact-num`]().
fn exact-num {|string-or-number| }

# Coerces the argument to an inexact number.
#
# If the argument is a string, it is converted to a typed number first. If the
# argument is already an inexact number, it is returned as is.
#
# Examples:
#
# ```elvish-transcript
# ~> inexact-num (num 1)
# ▶ (num 1.0)
# ~> inexact-num (num 0.5)
# ▶ (num 0.5)
# ~> inexact-num (num 1/2)
# ▶ (num 0.5)
# ~> inexact-num 1/2
# ▶ (num 0.5)
# ```
#
# Since the underlying representation for inexact numbers has limited range,
# numbers with very large magnitudes may be converted to an infinite value:
#
# ```elvish-transcript
# ~> inexact-num 1000000000000000000
# ▶ (num 1e+18)
# ~> inexact-num 10000000000000000000
# ▶ (num +Inf)
# ~> inexact-num -10000000000000000000
# ▶ (num -Inf)
# ```
#
# Likewise, numbers with very small magnitudes may be converted to 0:
#
# ```elvish-transcript
# ~> use math
# ~> math:pow 10 -323
# ▶ (num 1/100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000)
# ~> inexact-num (math:pow 10 -323)
# ▶ (num 1e-323)
# ~> math:pow 10 -324
# ▶ (num 1/1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000)
# ~> inexact-num (math:pow 10 -324)
# ▶ (num 0.0)
# ```
#
# See also [`num`]() and [`exact-num`]().
fn inexact-num {|string-or-number| }

#doc:html-id num-lt
# Outputs whether `$number`s in the given order are numerically strictly
# increasing. Outputs `$true` when given fewer than two numbers.
#
# Examples:
#
# ```elvish-transcript
# ~> < 1 2
# ▶ $true
# ~> < 2 1
# ▶ $false
# ~> < 1 2 3
# ▶ $true
# ```
#
fn '<' {|@number| }

#doc:html-id num-le
# Outputs whether `$number`s in the given order are numerically non-decreaing.
# Outputs `$true` when given fewer than two numbers.
#
# Examples:
#
# ```elvish-transcript
# ~> <= 1 1
# ▶ $true
# ~> <= 2 1
# ▶ $false
# ~> <= 1 1 2
# ▶ $true
# ```
fn '<=' {|@number| }

#doc:html-id num-eq
# Outputs whether `$number`s are all numerically equal. Outputs `$true` when
# given fewer than two numbers.
#
# Examples:
#
# ```elvish-transcript
# ~> == 1 1
# ▶ $true
# ~> == 1 (num 1)
# ▶ $true
# ~> == 1 (num 1) 1
# ▶ $true
# ~> == 1 (num 1) 1.0
# ▶ $true
# ~> == 1 2
# ▶ $false
# ```
fn '==' {|@number| }

#doc:html-id num-ne
# Determines whether `$a` and `$b` are numerically inequal. Equivalent to `not
# (== $a $b)`.
#
# Examples:
#
# ```elvish-transcript
# ~> != 1 2
# ▶ $true
# ~> != 1 1
# ▶ $false
# ```
fn '!=' {|a b| }

#doc:html-id num-gt
# Determines whether `$number`s in the given order are numerically strictly
# decreasing. Outputs `$true` when given fewer than two numbers.
#
# Examples:
#
# ```elvish-transcript
# ~> > 2 1
# ▶ $true
# ~> > 1 2
# ▶ $false
# ~> > 3 2 1
# ▶ $true
# ```
fn '>' {|@number| }

#doc:html-id num-ge
# Outputs whether `$number`s in the given order are numerically non-increasing.
# Outputs `$true` when given fewer than two numbers.
#
# Examples:
#
# ```elvish-transcript
# ~> >= 1 1
# ▶ $true
# ~> >= 1 2
# ▶ $false
# ~> >= 2 1 1
# ▶ $true
# ```
fn '>=' {|@number| }

#doc:html-id add
# Outputs the sum of all arguments, or 0 when there are no arguments.
#
# This command is [exactness-preserving](#exactness-preserving).
#
# Examples:
#
# ```elvish-transcript
# ~> + 5 2 7
# ▶ (num 14)
# ~> + 1/2 1/3 1/4
# ▶ (num 13/12)
# ~> + 1/2 0.5
# ▶ (num 1.0)
# ```
fn + {|@num| }

#doc:html-id sub
# Outputs the result of subtracting from `$x-num` all the `$y-num`s, working
# from left to right. When no `$y-num` is given, outputs the negation of
# `$x-num` instead (in other words, `- $x-num` is equivalent to `- 0 $x-num`).
#
# This command is [exactness-preserving](#exactness-preserving).
#
# Examples:
#
# ```elvish-transcript
# ~> - 5
# ▶ (num -5)
# ~> - 5 2
# ▶ (num 3)
# ~> - 5 2 7
# ▶ (num -4)
# ~> - 1/2 1/3
# ▶ (num 1/6)
# ~> - 1/2 0.3
# ▶ (num 0.2)
# ~> - 10
# ▶ (num -10)
# ```
fn - {|x-num @y-num| }

#doc:html-id mul
# Outputs the product of all arguments, or 1 when there are no arguments.
#
# This command is [exactness-preserving](#exactness-preserving). Additionally,
# when any argument is exact 0 and no other argument is a floating-point
# infinity, the result is exact 0.
#
# Examples:
#
# ```elvish-transcript
# ~> * 2 5 7
# ▶ (num 70)
# ~> * 1/2 0.5
# ▶ (num 0.25)
# ~> * 0 0.5
# ▶ (num 0)
# ```
fn * {|@num| }

#doc:html-id div
# Outputs the result of dividing `$x-num` with all the `$y-num`s, working from
# left to right. When no `$y-num` is given, outputs the reciprocal of `$x-num`
# instead (in other words, `/ $y-num` is equivalent to `/ 1 $y-num`).
#
# Dividing by exact 0 raises an exception. Dividing by inexact 0 results with
# either infinity or NaN according to floating-point semantics.
#
# This command is [exactness-preserving](#exactness-preserving). Additionally,
# when `$x-num` is exact 0 and no `$y-num` is exact 0, the result is exact 0.
#
# Examples:
#
# ```elvish-transcript
# ~> / 2
# ▶ (num 1/2)
# ~> / 2.0
# ▶ (num 0.5)
# ~> / 10 5
# ▶ (num 2)
# ~> / 2 5
# ▶ (num 2/5)
# ~> / 2 5 7
# ▶ (num 2/35)
# ~> / 0 1.0
# ▶ (num 0)
# ~> / 2 0
# Exception: bad value: divisor must be number other than exact 0, but is exact 0
#   [tty]:1:1-5: / 2 0
# ~> / 2 0.0
# ▶ (num +Inf)
# ```
#
# When given no argument, this command is equivalent to `cd /`, due to the
# implicit cd feature. (The implicit cd feature is deprecated since 0.21.0).
fn / {|x-num @y-num| }

#doc:html-id rem
# Outputs the remainder after dividing `$x` by `$y`. The result has the same
# sign as `$x`. Both arguments must be exact integers.
#
# Examples:
#
# ```elvish-transcript
# ~> % 10 3
# ▶ (num 1)
# ~> % -10 3
# ▶ (num -1)
# ~> % 10 -3
# ▶ (num 1)
# ~> % 10000000000000000000 3
# ▶ (num 1)
# ~> % 10.0 3
# Exception: bad value: argument must be exact integer, but is (num 10.0)
#   [tty]:1:1-8: % 10.0 3
# ```
#
# This limit may be lifted in the future.
fn % {|x y| }

#//skip-test
# Output a pseudo-random integer N such that `$low <= N < $high`. If not given,
# `$low` defaults to 0. Examples:
#
# ```elvish-transcript
# ~> # Emulate dice
# randint 1 7
# ▶ 6
# ```
fn randint {|low? high| }

#doc:show-unstable
# Sets the seed for the random number generator.
fn -randseed {|seed| }

# Outputs numbers, starting from `$start` and ending before `$end`, using
# `&step` as the increment.
#
# - If `$start` <= `$end`, `&step` defaults to 1, and `range` outputs values as
#   long as they are smaller than `$end`. An exception is thrown if `&step` is
#   given a negative value.
#
# - If `$start` > `$end`, `&step` defaults to -1, and `range` outputs values as
#   long as they are greater than `$end`. An exception is thrown if `&step` is
#   given a positive value.
#
# As a special case, if the outputs are floating point numbers, `range` also
# terminates if the values stop changing.
#
# This command is [exactness-preserving](#exactness-preserving).
#
# Examples:
#
# ```elvish-transcript
# ~> range 4
# ▶ (num 0)
# ▶ (num 1)
# ▶ (num 2)
# ▶ (num 3)
# ~> range 4 0
# ▶ (num 4)
# ▶ (num 3)
# ▶ (num 2)
# ▶ (num 1)
# ~> range -3 3 &step=2
# ▶ (num -3)
# ▶ (num -1)
# ▶ (num 1)
# ~> range 3 -3 &step=-2
# ▶ (num 3)
# ▶ (num 1)
# ▶ (num -1)
# ~> use math
# ~> range (- (math:pow 2 53) 1) +inf
# ▶ (num 9007199254740991.0)
# ▶ (num 9007199254740992.0)
# ```
#
# When using floating-point numbers, beware that numerical errors can result in
# an incorrect number of outputs:
#
# ```elvish-transcript
# ~> range 0.9 &step=0.3
# ▶ (num 0.0)
# ▶ (num 0.3)
# ▶ (num 0.6)
# ▶ (num 0.8999999999999999)
# ```
#
# Avoid this problem by using exact rationals:
#
# ```elvish-transcript
# ~> range 9/10 &step=3/10
# ▶ (num 0)
# ▶ (num 3/10)
# ▶ (num 3/5)
# ```
#
# One usage of this command is to execute something a fixed number of times by
# combining with [each](#each):
#
# ```elvish-transcript
# ~> range 3 | each {|_| echo foo }
# foo
# foo
# foo
# ```
#
# Etymology:
# [Python](https://docs.python.org/3/library/functions.html#func-range).
fn range {|&step start=0 end| }
