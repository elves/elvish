# ```elvish
# <s  $string... # less
# <=s $string... # less or equal
# ==s $string... # equal
# !=s $string... # not equal
# >s  $string... # greater
# >=s $string... # greater or equal
# ```
#
# String comparisons. They behave similarly to their number counterparts when
# given multiple arguments. Examples:
#
# ```elvish-transcript
# ~> >s lorem ipsum
# ▶ $true
# ~> ==s 1 1.0
# ▶ $false
# ~> >s 8 12
# ▶ $true
# ```
#doc:id str-cmp
#doc:fn <s <=s ==s !=s >s >=s

# Output the width of `$string` when displayed on the terminal. Examples:
#
# ```elvish-transcript
# ~> wcswidth a
# ▶ 1
# ~> wcswidth lorem
# ▶ 5
# ~> wcswidth 你好，世界
# ▶ 10
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

# For each [value input](#value-inputs), calls `$f` with the input followed by
# all its fields. A [`break`](./builtin.html#break) command will cause `eawk`
# to stop processing inputs. A [`continue`](./builtin.html#continue) command
# will exit $f, but is ignored by `eawk`.
#
# It should behave the same as the following functions:
#
# ```elvish
# fn eawk {|f @rest|
#   each {|line|
#     var @fields = (re:split '[ \t]+' (str:trim $line " \t"))
#     $f $line $@fields
#   } $@rest
# }
# ```
#
# This command allows you to write code very similar to `awk` scripts using
# anonymous functions. Example:
#
# ```elvish-transcript
# ~> echo " lorem ipsum\n1 2" | awk '{ print $1 }'
# lorem
# 1
# ~> echo " lorem ipsum\n1 2" | eawk {|line a b| put $a }
# ▶ lorem
# ▶ 1
# ```
#
# **Note**: Since Elvish allows variable names consisting solely of digits, you
# can also do the following:
#
# ```elvish-transcript
# ~> echo " lorem ipsum\n1 2" | eawk {|0 1 2| put $1 }
# ▶ lorem
# ▶ 1
# ```
fn eawk {|f inputs?| }
