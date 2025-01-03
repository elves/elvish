#doc:html-id str-lt
# Outputs whether `$string`s in the given order are strictly increasing. Outputs
# `$true` when given fewer than two strings.
fn '<s' {|@string| }

#doc:html-id str-le
# Outputs whether `$string`s in the given order are strictly non-decreasing.
# Outputs `$true` when given fewer than two strings.
fn '<=s' {|@string| }

#doc:html-id str-eq
# Outputs whether `$string`s are all the same string. Outputs `$true` when given
# fewer than two strings.
fn '==s' {|@string| }

#doc:html-id str-ne
# Outputs whether `$a` and `$b` are not the same string. Equivalent to `not (==s
# $a $b)`.
fn '!=s' {|a b| }

#doc:html-id str-gt
# Outputs whether `$string`s in the given order are strictly decreasing. Outputs
# `$true` when given fewer than two strings.
fn '>s' {|@string| }

#doc:html-id str-ge
# Outputs whether `$string`s in the given order are strictly non-increasing.
# Outputs `$true` when given fewer than two strings.
fn '>=s' {|@string| }

# Output the width of `$string` when displayed on the terminal. Examples:
#
# ```elvish-transcript
# ~> wcswidth a
# ▶ (num 1)
# ~> wcswidth lorem
# ▶ (num 5)
# ~> wcswidth 你好，世界
# ▶ (num 10)
# ```
fn wcswidth {|string| }

# Convert arguments to string values.
#
# ```elvish-transcript
# ~> to-string foo [a] [&k=v]
# ▶ foo
# ▶ '[a]'
# ▶ '[&k=v]'
# ```
fn to-string {|@value| }

# Outputs a string for each `$number` written in `$base`. The `$base` must be
# between 2 and 36, inclusive. Examples:
#
# ```elvish-transcript
# ~> base 2 1 3 4 16 255
# ▶ 1
# ▶ 11
# ▶ 100
# ▶ 10000
# ▶ 11111111
# ~> base 16 1 3 4 16 255
# ▶ 1
# ▶ 3
# ▶ 4
# ▶ 10
# ▶ ff
# ```
fn base {|base @number| }
