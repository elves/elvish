#//each:eval use re

# Quote `$string` for use in a pattern. Examples:
#
# ```elvish-transcript
# ~> re:quote a.txt
# ▶ a\.txt
# ~> re:quote '(*)'
# ▶ '\(\*\)'
# ```
fn quote {|string| }

# Determine whether `$pattern` matches `$source`. The pattern is not anchored.
# Examples:
#
# ```elvish-transcript
# ~> re:match . xyz
# ▶ $true
# ~> re:match . ''
# ▶ $false
# ~> re:match '[a-z]' A
# ▶ $false
# ```
fn match {|&posix=$false pattern source| }

# Find all matches of `$pattern` in `$source`.
#
# Each match is represented by a map-like value `$m`; `$m[text]`, `$m[start]` and
# `$m[end]` are the text, start and end positions (as byte indices into `$source`)
# of the match; `$m[groups]` is a list of submatches for capture groups in the
# pattern. A submatch has a similar structure to a match, except that it does not
# have a `group` key. The entire pattern is an implicit capture group, and it
# always appears first.
#
# Examples:
#
# ```elvish-transcript
# ~> re:find . ab
# ▶ [&end=(num 1) &groups=[[&end=(num 1) &start=(num 0) &text=a]] &start=(num 0) &text=a]
# ▶ [&end=(num 2) &groups=[[&end=(num 2) &start=(num 1) &text=b]] &start=(num 1) &text=b]
# ~> re:find '[A-Z]([0-9])' 'A1 B2'
# ▶ [&end=(num 2) &groups=[[&end=(num 2) &start=(num 0) &text=A1] [&end=(num 2) &start=(num 1) &text=1]] &start=(num 0) &text=A1]
# ▶ [&end=(num 5) &groups=[[&end=(num 5) &start=(num 3) &text=B2] [&end=(num 5) &start=(num 4) &text=2]] &start=(num 3) &text=B2]
# ```
fn find {|&posix=$false &longest=$false &max=-1 pattern source| }

# Replace all occurrences of `$pattern` in `$source` with `$repl`.
#
# The replacement `$repl` can be any of the following:
#
# -   A string-typed replacement template. The template can use `$name` or
#     `${name}` patterns to refer to capture groups, where `name` consists of
#     letters, digits and underscores. A purely numeric patterns like `$1`
#     refers to the capture group with the corresponding index; other names
#     refer to capture groups named with the `(?P<name>...)`) syntax.
#
#     In the `$name` form, the name is taken to be as long as possible; `$1` is
#     equivalent to `${1x}`, not `${1}x`; `$10` is equivalent to `${10}`, not `${1}0`.
#
#     To insert a literal `$`, use `$$`.
#
# -   A function that takes a string argument and outputs a string. For each
#     match, the function is called with the content of the match, and its output
#     is used as the replacement.
#
# If `$literal` is true, `$repl` must be a string and is treated literally instead
# of as a pattern.
#
# Example:
#
# ```elvish-transcript
# ~> re:replace '(ba|z)sh' '${1}SH' 'bash and zsh'
# ▶ 'baSH and zSH'
# ~> re:replace '(ba|z)sh' elvish 'bash and zsh rock'
# ▶ 'elvish and elvish rock'
# ~> re:replace '(ba|z)sh' {|x| put [&bash=BaSh &zsh=ZsH][$x] } 'bash and zsh'
# ▶ 'BaSh and ZsH'
# ```
fn replace {|&posix=$false &longest=$false &literal=$false pattern repl source| }

# Split `$source`, using `$pattern` as separators. Examples:
#
# ```elvish-transcript
# ~> re:split : /usr/sbin:/usr/bin:/bin
# ▶ /usr/sbin
# ▶ /usr/bin
# ▶ /bin
# ~> re:split &max=2 : /usr/sbin:/usr/bin:/bin
# ▶ /usr/sbin
# ▶ /usr/bin:/bin
# ```
fn split {|&posix=$false &longest=$false &max=-1 pattern source| }

# For each [value input](builtin.html#value-inputs), calls `$f` with the input
# followed by all its fields.
#
# The `&sep` option is a regular expression for the field separator. For the
# `&sep-posix` and `&sep-longest` options, see the
# [introduction](#introduction); the `sep-` prefix is added for clarity.
#
# Calling [`break`]() in `$f` exits both `$f` and `re:awk`, and can be used to
# stop processing inputs early. Calling [`continue`]() exits `$f` but not
# `re:awk`, and can be used to stop `$f` early but continue processing inputs.
#
# This command allows you to write code resembling
# [AWK](https://en.wikipedia.org/wiki/AWK) scripts, using an anonymous function
# instead of a string containing AWK code. A simple example:
#
# ```elvish-transcript
# ~> echo " lorem ipsum\n1 2" | awk '{ print $1 }'
# lorem
# 1
# ~> echo " lorem ipsum\n1 2" | re:awk {|line a b| put $a }
# ▶ lorem
# ▶ 1
# ```
#
# **Note**: Since Elvish allows variable names consisting solely of digits, you
# can do something like this to emulate AWK even more closely:
#
# ```elvish-transcript
# ~> echo " lorem ipsum\n1 2" | re:awk {|0 1 2| put $1 }
# ▶ lorem
# ▶ 1
# ```
#
# If the number of fields differ between lines, use a rest argument:
#
# ```elvish-transcript
# ~> echo "a b\nc d e" | re:awk {|@a| echo (- (count $a) 1)' fields' }
# 2 fields
# 3 fields
# ```
#
# This command is roughly equivalent to the following Elvish function:
#
# ```elvish
# fn my-awk {|&sep='[ \t]+' &sep-posix=$false &sep-longest=$false f @rest|
#   each {|line|
#     var @fields = (re:split $sep &posix=$sep-posix &longest=$sep-longest (str:trim $line " \t"))
#     $f $line $@fields
#   } $@rest
# }
# ```
fn awk {|&sep='[ \t]+' &sep-posix=$false &sep-longest=$false f inputs?| }
