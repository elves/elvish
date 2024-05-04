#//each:eval use str

# Compares two strings and output an integer that will be 0 if a == b,
# -1 if a < b, and +1 if a > b.
#
# ```elvish-transcript
# ~> str:compare a a
# ▶ (num 0)
# ~> str:compare a b
# ▶ (num -1)
# ~> str:compare b a
# ▶ (num 1)
# ```
fn compare {|a b| }

# Outputs whether `$str` contains `$substr` as a substring.
#
# ```elvish-transcript
# ~> str:contains abcd x
# ▶ $false
# ~> str:contains abcd bc
# ▶ $true
# ```
fn contains {|str substr| }

# Outputs whether `$str` contains any Unicode code points in `$chars`.
#
# ```elvish-transcript
# ~> str:contains-any abcd x
# ▶ $false
# ~> str:contains-any abcd xby
# ▶ $true
# ```
fn contains-any {|str chars| }

# Outputs the number of non-overlapping instances of `$substr` in `$s`.
# If `$substr` is an empty string, output 1 + the number of Unicode code
# points in `$s`.
#
# ```elvish-transcript
# ~> str:count abcdefabcdef bc
# ▶ (num 2)
# ~> str:count abcdef ''
# ▶ (num 7)
# ```
fn count {|str substr| }

# Outputs if `$str1` and `$str2`, interpreted as UTF-8 strings, are equal
# under Unicode case-folding.
#
# ```elvish-transcript
# ~> str:equal-fold ABC abc
# ▶ $true
# ~> str:equal-fold abc ab
# ▶ $false
# ```
fn equal-fold {|str1 str2| }


# Splits `$str` around each instance of one or more consecutive white space
# characters.
#
# ```elvish-transcript
# ~> str:fields "lorem ipsum   dolor"
# ▶ lorem
# ▶ ipsum
# ▶ dolor
# ~> str:fields "   "
# ```
#
# See also [`str:split`]().
fn fields {|str| }

# Outputs a string consisting of the given Unicode codepoints. Example:
#
# ```elvish-transcript
# ~> str:from-codepoints 0x61
# ▶ a
# ~> str:from-codepoints 0x4f60 0x597d
# ▶ 你好
# ```
#
# See also [`str:to-codepoints`]().
fn from-codepoints {|@number| }

# Outputs a string consisting of the given Unicode bytes. Example:
#
# ```elvish-transcript
# ~> str:from-utf8-bytes 0x61
# ▶ a
# ~> str:from-utf8-bytes 0xe4 0xbd 0xa0 0xe5 0xa5 0xbd
# ▶ 你好
# ```
#
# See also [`str:to-utf8-bytes`]().
fn from-utf8-bytes {|@number| }

# Outputs if `$str` begins with `$prefix`.
#
# ```elvish-transcript
# ~> str:has-prefix abc ab
# ▶ $true
# ~> str:has-prefix abc bc
# ▶ $false
# ```
fn has-prefix {|str prefix| }

# Outputs if `$str` ends with `$suffix`.
#
# ```elvish-transcript
# ~> str:has-suffix abc ab
# ▶ $false
# ~> str:has-suffix abc bc
# ▶ $true
# ```
fn has-suffix {|str suffix| }

# Outputs the index of the first instance of `$substr` in `$str`, or -1
# if `$substr` is not present in `$str`.
#
# ```elvish-transcript
# ~> str:index abcd cd
# ▶ (num 2)
# ~> str:index abcd xyz
# ▶ (num -1)
# ```
fn index {|str substr| }

# Outputs the index of the first instance of any Unicode code point
# from `$chars` in `$str`, or -1 if no Unicode code point from `$chars` is
# present in `$str`.
#
# ```elvish-transcript
# ~> str:index-any "chicken" "aeiouy"
# ▶ (num 2)
# ~> str:index-any l33t aeiouy
# ▶ (num -1)
# ```
fn index-any {|str chars| }

# Joins inputs with `$sep`. Examples:
#
# ```elvish-transcript
# ~> put lorem ipsum | str:join ,
# ▶ 'lorem,ipsum'
# ~> str:join , [lorem ipsum]
# ▶ 'lorem,ipsum'
# ~> str:join '' [lorem ipsum]
# ▶ loremipsum
# ~> str:join '...' [lorem ipsum]
# ▶ lorem...ipsum
# ```
#
# Etymology: Various languages,
# [Python](https://docs.python.org/3.6/library/stdtypes.html#str.join).
#
# See also [`str:split`]().
fn join {|sep input-list?| }

# Outputs the index of the last instance of `$substr` in `$str`,
# or -1 if `$substr` is not present in `$str`.
#
# ```elvish-transcript
# ~> str:last-index "elven speak elvish" elv
# ▶ (num 12)
# ~> str:last-index "elven speak elvish" romulan
# ▶ (num -1)
# ```
fn last-index {|str substr| }

#doc:added-in 0.21
# Outputs a string consisting of `$n` copies of `$s`.
#
# Examples:
#
# ```elvish-transcript
# ~> str:repeat a 10
# ▶ aaaaaaaaaa
# ```
fn repeat {|s n| }

# Replaces all occurrences of `$old` with `$repl` in `$source`. If `$max` is
# non-negative, it determines the max number of substitutions.
#
# **Note**: This command does not support searching by regular expressions, `$old`
# is always interpreted as a plain string. Use [re:replace](re.html#re:replace) if
# you need to search by regex.
fn replace {|&max=-1 old repl source| }

# Splits `$string` by `$sep`. If `$sep` is an empty string, split it into
# codepoints.
#
# If the `&max` option is non-negative, stops after producing the maximum
# number of results.
#
# ```elvish-transcript
# ~> str:split , lorem,ipsum
# ▶ lorem
# ▶ ipsum
# ~> str:split '' 你好
# ▶ 你
# ▶ 好
# ~> str:split &max=2 ' ' 'a b c d'
# ▶ a
# ▶ 'b c d'
# ```
#
# **Note**: This command does not support splitting by regular expressions,
# `$sep` is always interpreted as a plain string. Use [re:split](re.html#re:split)
# if you need to split by regex.
#
# Etymology: Various languages, in particular
# [Python](https://docs.python.org/3.6/library/stdtypes.html#str.split).
#
# See also [`str:join`]() and [`str:fields`]().
fn split {|&max=-1 sep string| }

# Outputs `$str` with all Unicode letters that begin words mapped to their
# Unicode title case.
#
# ```elvish-transcript
# ~> str:title "her royal highness"
# ▶ 'Her Royal Highness'
# ```
fn title {|str| }

# Outputs value of each codepoint in `$string`, in hexadecimal. Examples:
#
# ```elvish-transcript
# ~> str:to-codepoints a
# ▶ 0x61
# ~> str:to-codepoints 你好
# ▶ 0x4f60
# ▶ 0x597d
# ```
#
# The output format is subject to change.
#
# See also [`str:from-codepoints`]().
fn to-codepoints {|string| }

# Outputs `$str` with all Unicode letters mapped to their lower-case
# equivalent.
#
# ```elvish-transcript
# ~> str:to-lower 'ABC!123'
# ▶ abc!123
# ```
fn to-lower {|str| }

# Outputs value of each byte in `$string`, in hexadecimal. Examples:
#
# ```elvish-transcript
# ~> str:to-utf8-bytes a
# ▶ 0x61
# ~> str:to-utf8-bytes 你好
# ▶ 0xe4
# ▶ 0xbd
# ▶ 0xa0
# ▶ 0xe5
# ▶ 0xa5
# ▶ 0xbd
# ```
#
# The output format is subject to change.
#
# See also [`str:from-utf8-bytes`]().
fn to-utf8-bytes {|string| }

# Outputs `$str` with all Unicode letters mapped to their Unicode title case.
#
# ```elvish-transcript
# ~> str:to-title "her royal highness"
# ▶ 'HER ROYAL HIGHNESS'
# ~> str:to-title "хлеб"
# ▶ ХЛЕБ
# ```
fn to-title {|str| }

# Outputs `$str` with all Unicode letters mapped to their upper-case
# equivalent.
#
# ```elvish-transcript
# ~> str:to-upper 'abc!123'
# ▶ ABC!123
# ```
fn to-upper { }

# Outputs `$str` with all leading and trailing Unicode code points contained
# in `$cutset` removed.
#
# ```elvish-transcript
# ~> str:trim "¡¡¡Hello, Elven!!!" "!¡"
# ▶ 'Hello, Elven'
# ```
fn trim {|str cutset| }

# Outputs `$str` with all leading Unicode code points contained in `$cutset`
# removed. To remove a prefix string use [`str:trim-prefix`]().
#
# ```elvish-transcript
# ~> str:trim-left "¡¡¡Hello, Elven!!!" "!¡"
# ▶ 'Hello, Elven!!!'
# ```
fn trim-left {|str cutset| }

# Outputs `$str` minus the leading `$prefix` string. If `$str` doesn't begin
# with `$prefix`, `$str` is output unchanged.
#
# ```elvish-transcript
# ~> str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hello, "
# ▶ Elven!!!
# ~> str:trim-prefix "¡¡¡Hello, Elven!!!" "¡¡¡Hola, "
# ▶ '¡¡¡Hello, Elven!!!'
# ```
fn trim-prefix {|str prefix| }

# Outputs `$str` with all trailing Unicode code points contained in `$cutset`
# removed. To remove a suffix string use [`str:trim-suffix`]().
#
# ```elvish-transcript
# ~> str:trim-right "¡¡¡Hello, Elven!!!" "!¡"
# ▶ '¡¡¡Hello, Elven'
# ```
fn trim-right {|str cutset| }

# Outputs `$str` with all leading and trailing white space removed as defined
# by Unicode.
#
# ```elvish-transcript
# ~> str:trim-space " \t\n Hello, Elven \n\t\r\n"
# ▶ 'Hello, Elven'
# ```
fn trim-space {|str| }

# Outputs `$str` minus the trailing `$suffix` string. If `$str` doesn't end
# with `$suffix`, `$str` is output unchanged.
#
# ```elvish-transcript
# ~> str:trim-suffix "¡¡¡Hello, Elven!!!" ", Elven!!!"
# ▶ ¡¡¡Hello
# ~> str:trim-suffix "¡¡¡Hello, Elven!!!" ", Klingons!!!"
# ▶ '¡¡¡Hello, Elven!!!'
# ```
fn trim-suffix {|str suffix| }
