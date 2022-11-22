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
# ▶ [&text=a &start=0 &end=1 &groups=[[&text=a &start=0 &end=1]]]
# ▶ [&text=b &start=1 &end=2 &groups=[[&text=b &start=1 &end=2]]]
# ~> re:find '[A-Z]([0-9])' 'A1 B2'
# ▶ [&text=A1 &start=0 &end=2 &groups=[[&text=A1 &start=0 &end=2] [&text=1 &start=1 &end=2]]]
# ▶ [&text=B2 &start=3 &end=5 &groups=[[&text=B2 &start=3 &end=5] [&text=2 &start=4 &end=5]]]
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
