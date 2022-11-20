#elvdoc:var e
#
# Approximate value of
# [`e`](https://en.wikipedia.org/wiki/E_(mathematical_constant)):
# 2.718281.... This variable is read-only.

#elvdoc:var pi
#
# Approximate value of [`π`](https://en.wikipedia.org/wiki/Pi): 3.141592.... This
# variable is read-only.

#elvdoc:fn abs
#
# ```elvish
# math:abs $number
# ```
#
# Computes the absolute value `$number`. This function is exactness-preserving.
# Examples:
#
# ```elvish-transcript
# ~> math:abs 2
# ▶ (num 2)
# ~> math:abs -2
# ▶ (num 2)
# ~> math:abs 10000000000000000000
# ▶ (num 10000000000000000000)
# ~> math:abs -10000000000000000000
# ▶ (num 10000000000000000000)
# ~> math:abs 1/2
# ▶ (num 1/2)
# ~> math:abs -1/2
# ▶ (num 1/2)
# ~> math:abs 1.23
# ▶ (num 1.23)
# ~> math:abs -1.23
# ▶ (num 1.23)
# ```

#elvdoc:fn acos
#
# ```elvish
# math:acos $number
# ```
#
# Outputs the arccosine of `$number`, in radians (not degrees). Examples:
#
# ```elvish-transcript
# ~> math:acos 1
# ▶ (num 1)
# ~> math:acos 1.00001
# ▶ (num NaN)
# ```

#elvdoc:fn acosh
#
# ```elvish
# math:acosh $number
# ```
#
# Outputs the inverse hyperbolic cosine of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:acosh 1
# ▶ (num 0)
# ~> math:acosh 0
# ▶ (num NaN)
# ```

#elvdoc:fn asin
#
# ```elvish
# math:asin $number
# ```
#
# Outputs the arcsine of `$number`, in radians (not degrees). Examples:
#
# ```elvish-transcript
# ~> math:asin 0
# ▶ (num 0)
# ~> math:asin 1
# ▶ (num 1.5707963267948966)
# ~> math:asin 1.00001
# ▶ (num NaN)
# ```

#elvdoc:fn asinh
#
# ```elvish
# math:asinh $number
# ```
#
# Outputs the inverse hyperbolic sine of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:asinh 0
# ▶ (num 0)
# ~> math:asinh inf
# ▶ (num +Inf)
# ```

#elvdoc:fn atan
#
# ```elvish
# math:atan $number
# ```
#
# Outputs the arctangent of `$number`, in radians (not degrees). Examples:
#
# ```elvish-transcript
# ~> math:atan 0
# ▶ (num 0)
# ~> math:atan $math:inf
# ▶ (num 1.5707963267948966)
# ```

#elvdoc:fn atanh
#
# ```elvish
# math:atanh $number
# ```
#
# Outputs the inverse hyperbolic tangent of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:atanh 0
# ▶ (num 0)
# ~> math:atanh 1
# ▶ (num +Inf)
# ```

#elvdoc:fn ceil
#
# ```elvish
# math:ceil $number
# ```
#
# Computes the least integer greater than or equal to `$number`. This function
# is exactness-preserving.
#
# The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
# NaN are themselves.
#
# Examples:
#
# ```elvish-transcript
# ~> math:floor 1
# ▶ (num 1)
# ~> math:floor 3/2
# ▶ (num 1)
# ~> math:floor -3/2
# ▶ (num -2)
# ~> math:floor 1.1
# ▶ (num 1.0)
# ~> math:floor -1.1
# ▶ (num -2.0)
# ```

#elvdoc:fn cos
#
# ```elvish
# math:cos $number
# ```
#
# Computes the cosine of `$number` in units of radians (not degrees).
# Examples:
#
# ```elvish-transcript
# ~> math:cos 0
# ▶ (num 1)
# ~> math:cos 3.14159265
# ▶ (num -1)
# ```

#elvdoc:fn cosh
#
# ```elvish
# math:cosh $number
# ```
#
# Computes the hyperbolic cosine of `$number`. Example:
#
# ```elvish-transcript
# ~> math:cosh 0
# ▶ (num 1)
# ```

#elvdoc:fn floor
#
# ```elvish
# math:floor $number
# ```
#
# Computes the greatest integer less than or equal to `$number`. This function
# is exactness-preserving.
#
# The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
# NaN are themselves.
#
# Examples:
#
# ```elvish-transcript
# ~> math:floor 1
# ▶ (num 1)
# ~> math:floor 3/2
# ▶ (num 1)
# ~> math:floor -3/2
# ▶ (num -2)
# ~> math:floor 1.1
# ▶ (num 1.0)
# ~> math:floor -1.1
# ▶ (num -2.0)
# ```

#elvdoc:fn is-inf
#
# ```elvish
# math:is-inf &sign=0 $number
# ```
#
# Tests whether the number is infinity. If sign > 0, tests whether `$number`
# is positive infinity. If sign < 0, tests whether `$number` is negative
# infinity. If sign == 0, tests whether `$number` is either infinity.
#
# ```elvish-transcript
# ~> math:is-inf 123
# ▶ $false
# ~> math:is-inf inf
# ▶ $true
# ~> math:is-inf -inf
# ▶ $true
# ~> math:is-inf &sign=1 inf
# ▶ $true
# ~> math:is-inf &sign=-1 inf
# ▶ $false
# ~> math:is-inf &sign=-1 -inf
# ▶ $true
# ```

#elvdoc:fn is-nan
#
# ```elvish
# math:is-nan $number
# ```
#
# Tests whether the number is a NaN (not-a-number).
#
# ```elvish-transcript
# ~> math:is-nan 123
# ▶ $false
# ~> math:is-nan (num inf)
# ▶ $false
# ~> math:is-nan (num nan)
# ▶ $true
# ```

#elvdoc:fn log
#
# ```elvish
# math:log $number
# ```
#
# Computes the natural (base *e*) logarithm of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:log 1.0
# ▶ (num 1)
# ~> math:log -2.3
# ▶ (num NaN)
# ```

#elvdoc:fn log10
#
# ```elvish
# math:log10 $number
# ```
#
# Computes the base 10 logarithm of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:log10 100.0
# ▶ (num 2)
# ~> math:log10 -1.7
# ▶ (num NaN)
# ```

#elvdoc:fn log2
#
# ```elvish
# math:log2 $number
# ```
#
# Computes the base 2 logarithm of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:log2 8
# ▶ (num 3)
# ~> math:log2 -5.3
# ▶ (num NaN)
# ```

#elvdoc:fn max
#
# ```elvish
# math:max $number...
# ```
#
# Outputs the maximum number in the arguments. If there are no arguments,
# an exception is thrown. If any number is NaN then NaN is output. This
# function is exactness-preserving.
#
# Examples:
#
# ```elvish-transcript
# ~> math:max 3 5 2
# ▶ (num 5)
# ~> math:max (range 100)
# ▶ (num 99)
# ~> math:max 1/2 1/3 2/3
# ▶ (num 2/3)
# ```

#elvdoc:fn min
#
# ```elvish
# math:min $number...
# ```
#
# Outputs the minimum number in the arguments. If there are no arguments
# an exception is thrown. If any number is NaN then NaN is output. This
# function is exactness-preserving.
#
# Examples:
#
# ```elvish-transcript
# ~> math:min
# Exception: arity mismatch: arguments must be 1 or more values, but is 0 values
# [tty 17], line 1: math:min
# ~> math:min 3 5 2
# ▶ (num 2)
# ~> math:min 1/2 1/3 2/3
# ▶ (num 1/3)
# ```

#elvdoc:fn pow
#
# ```elvish
# math:pow $base $exponent
# ```
#
# Outputs the result of raising `$base` to the power of `$exponent`.
#
# This function produces an exact result when `$base` is exact and `$exponent`
# is an exact integer. Otherwise it produces an inexact result.
#
# Examples:
#
# ```elvish-transcript
# ~> math:pow 3 2
# ▶ (num 9)
# ~> math:pow -2 2
# ▶ (num 4)
# ~> math:pow 1/2 3
# ▶ (num 1/8)
# ~> math:pow 1/2 -3
# ▶ (num 8)
# ~> math:pow 9 1/2
# ▶ (num 3.0)
# ~> math:pow 12 1.1
# ▶ (num 15.38506624784179)
# ```

#elvdoc:fn round
#
# ```elvish
# math:round $number
# ```
#
# Outputs the nearest integer, rounding half away from zero. This function is
# exactness-preserving.
#
# The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
# NaN are themselves.
#
# Examples:
#
# ```elvish-transcript
# ~> math:round 2
# ▶ (num 2)
# ~> math:round 1/3
# ▶ (num 0)
# ~> math:round 1/2
# ▶ (num 1)
# ~> math:round 2/3
# ▶ (num 1)
# ~> math:round -1/3
# ▶ (num 0)
# ~> math:round -1/2
# ▶ (num -1)
# ~> math:round -2/3
# ▶ (num -1)
# ~> math:round 2.5
# ▶ (num 3.0)
# ```

#elvdoc:fn round-to-even
#
# ```elvish
# math:round-to-even $number
# ```
#
# Outputs the nearest integer, rounding ties to even. This function is
# exactness-preserving.
#
# The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
# NaN are themselves.
#
# Examples:
#
# ```elvish-transcript
# ~> math:round-to-even 2
# ▶ (num 2)
# ~> math:round-to-even 1/2
# ▶ (num 0)
# ~> math:round-to-even 3/2
# ▶ (num 2)
# ~> math:round-to-even 5/2
# ▶ (num 2)
# ~> math:round-to-even -5/2
# ▶ (num -2)
# ~> math:round-to-even 2.5
# ▶ (num 2.0)
# ~> math:round-to-even 1.5
# ▶ (num 2.0)
# ```

#elvdoc:fn sin
#
# ```elvish
# math:sin $number
# ```
#
# Computes the sine of `$number` in units of radians (not degrees). Examples:
#
# ```elvish-transcript
# ~> math:sin 0
# ▶ (num 0)
# ~> math:sin 3.14159265
# ▶ (num 3.5897930298416118e-09)
# ```

#elvdoc:fn sinh
#
# ```elvish
# math:sinh $number
# ```
#
# Computes the hyperbolic sine of `$number`. Example:
#
# ```elvish-transcript
# ~> math:sinh 0
# ▶ (num 0)
# ```

#elvdoc:fn sqrt
#
# ```elvish
# math:sqrt $number
# ```
#
# Computes the square-root of `$number`. Examples:
#
# ```elvish-transcript
# ~> math:sqrt 0
# ▶ (num 0)
# ~> math:sqrt 4
# ▶ (num 2)
# ~> math:sqrt -4
# ▶ (num NaN)
# ```

#elvdoc:fn tan
#
# ```elvish
# math:tan $number
# ```
#
# Computes the tangent of `$number` in units of radians (not degrees). Examples:
#
# ```elvish-transcript
# ~> math:tan 0
# ▶ (num 0)
# ~> math:tan 3.14159265
# ▶ (num -0.0000000035897930298416118)
# ```

#elvdoc:fn tanh
#
# ```elvish
# math:tanh $number
# ```
#
# Computes the hyperbolic tangent of `$number`. Example:
#
# ```elvish-transcript
# ~> math:tanh 0
# ▶ (num 0)
# ```

#elvdoc:fn trunc
#
# ```elvish
# math:trunc $number
# ```
#
# Outputs the integer portion of `$number`. This function is exactness-preserving.
#
# The results for the special floating-point values -0.0, +0.0, -Inf, +Inf and
# NaN are themselves.
#
# Examples:
#
# ```elvish-transcript
# ~> math:trunc 1
# ▶ (num 1)
# ~> math:trunc 3/2
# ▶ (num 1)
# ~> math:trunc 5/3
# ▶ (num 1)
# ~> math:trunc -3/2
# ▶ (num -1)
# ~> math:trunc -5/3
# ▶ (num -1)
# ~> math:trunc 1.7
# ▶ (num 1.0)
# ~> math:trunc -1.7
# ▶ (num -1.0)
# ```
