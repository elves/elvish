# Takes arbitrary arguments and write them to the structured stdout.
#
# Examples:
#
# ```elvish-transcript
# ~> put a
# ▶ a
# ~> put lorem ipsum [a b] [&k=v]
# ▶ lorem
# ▶ ipsum
# ▶ [a b]
# ▶ [&k=v]
# ```
#
# **Note**: It is almost never necessary to use `put (...)` - just write the
# `...` part. For example, `put (eq a b)` is the equivalent to just `eq a b`.
#
# Etymology: Various languages, in particular
# [C](https://manpages.debian.org/stretch/manpages-dev/puts.3.en.html) and
# [Ruby](https://ruby-doc.org/core-2.2.2/IO.html#method-i-puts) as `puts`.
fn put {|@value| }

# Output `$value` for `$n` times. Example:
#
# ```elvish-transcript
# ~> repeat 0 lorem
# ~> repeat 4 NAN
# ▶ NAN
# ▶ NAN
# ▶ NAN
# ▶ NAN
# ```
#
# Etymology: [Clojure](https://clojuredocs.org/clojure.core/repeat).
fn repeat {|n value| }

# Reads `$n` bytes, or until end-of-file, and outputs the bytes as a string
# value. The result may not be a valid UTF-8 string.
#
# Examples:
#
# ```elvish-transcript
# ~> echo "a,b" | read-bytes 2
# ▶ 'a,'
# ~> echo "a,b" | read-bytes 10
# ▶ "a,b\n"
# ```
fn read-bytes {|n| }

# Reads byte input until `$terminator` or end-of-file is encountered. It outputs the part of the
# input read as a string value. The output contains the trailing `$terminator`, unless `read-upto`
# terminated at end-of-file.
#
# The `$terminator` must be a single ASCII character such as `"\x00"` (NUL).
#
# Examples:
#
# ```elvish-transcript
# ~> echo "a,b,c" | read-upto ","
# ▶ 'a,'
# ~> echo "foo\nbar" | read-upto "\n"
# ▶ "foo\n"
# ~> echo "a.elv\x00b.elv" | read-upto "\x00"
# ▶ "a.elv\x00"
# ~> print "foobar" | read-upto "\n"
# ▶ foobar
# ```
fn read-upto {|terminator| }

# Reads a single line from byte input, and writes the line to the value output,
# stripping the line ending. A line can end with `"\r\n"`, `"\n"`, or end of
# file. Examples:
#
# ```elvish-transcript
# ~> print line | read-line
# ▶ line
# ~> print "line\n" | read-line
# ▶ line
# ~> print "line\r\n" | read-line
# ▶ line
# ~> print "line-with-extra-cr\r\r\n" | read-line
# ▶ "line-with-extra-cr\r"
# ```
fn read-line { }

# Like `echo`, just without the newline.
#
# See also [`echo`]().
#
# Etymology: Various languages, in particular
# [Perl](https://perldoc.perl.org/functions/print.html) and
# [zsh](http://zsh.sourceforge.net/Doc/Release/Shell-Builtin-Commands.html), whose
# `print`s do not print a trailing newline.
fn print {|&sep=' ' @value| }

# Prints values to the byte stream according to a template.
#
# The template may contain "formatting verbs", sequences that start with `%`.
# Each verb corresponds to an argument, and specifies how that argument will be
# converted and formatted:
#
# - `%s`, `%q` and `%v` convert the argument to a string:
#
#     - `%s`: use [to-string](#to-string)
#     - `%q`: use [repr](#repr)
#     - `%v`: equivalent to `%s`
#     - `%#v`: equivalent to `%q`.
#
# - `%t` converts the argument to a boolean using [bool](#bool), and prints it
#   as either `true` or `false`.
#
# - `%b`, `%c`, `%d`, `%o`, `%O`, `%x`, `%X` and `%U` convert the argument to an
#   integer, and then print it as follows:
#
#   - `%b`: base 2
#   - `%c`: the character represented by the Unicode codepoint
#   - `%d`: base 10
#   - `%o`: base 8
#   - `%O`: base 8, with `0o` prefix
#   - `%x`: base 16, with lower-case letters for a-f
#   - `%X`: base 16, with upper-case letters for A-F
#   - `%U`: Unicode format like U+1234; same as `U+%04X`
#
# - `%e`, `%E`, `%f`, `%F`, `%g` and `%G` convert the argument to a
#   floating-point number using [`inexact-num`](), and print it as follows:
#
#   - `%e`: scientific notation like -1.234456e+78
#   - `%E`: scientific notation like -1.234456E+78
#   - `%f`: decimal point but no exponent, like 123.456
#   - `%F`: same as `%f`
#   - `%g`: `%e` for large exponents, `%f` otherwise
#   - `%G`: `%E` for large exponents, `%F` otherwise
#
# - `%%` prints a literal `%` and consumes no argument.
#
# Unsupported verbs not documented don't cause exceptions, but the output will
# contain content like `%!Z(unsupported formatting verb)`.
#
# Flags may be added between `%` and the verb to modify behavior like precision
# and padding. See Go's [`fmt`](https://golang.org/pkg/fmt/#hdr-Printing)
# package for details.
#
# This command does not add a trailing newline. Include it explicitly in the
# template, like `printf "Hello %s\n" world`.
#
# Examples:
#
# ```elvish-transcript
# ~> use math
# ~> printf ": %10s %.3f\n" Pi $math:pi
#    printf ": %10s %.3f\n" E $math:e
# :         Pi 3.142
# :          E 2.718
# ~> printf "%-10s %.2f %s\n" Pi $math:pi $math:pi
# Pi         3.14 3.141592653589793
# ~> printf "%d\n" 0b11100111
# 231
# ~> printf "%08b\n" 231
# 11100111
# ~> printf "list is: %q\n" [foo bar 'foo bar']
# list is: [foo bar 'foo bar']
# ```
#
# Since `printf` writes to the byte stream, capturing its output will generate
# one value per line. To capture it into one string value, use it together with
# [`slurp`](), as in `var s = (printf ... | slurp)`.
#
# **Note**: Compared to the [POSIX `printf`
# command](https://pubs.opengroup.org/onlinepubs/007908799/xcu/printf.html)
# found in other shells, there are 3 key differences:
#
# - The behavior of the formatting verbs are based on Go's
#   [`fmt`](https://golang.org/pkg/fmt/) package instead of the POSIX
#   specification.
#
# - The number of arguments after the formatting template must match the number
#   of formatting verbs. The POSIX command will repeat the template string to
#   consume excess values; this command does not have that behavior.
#
# - This command does not interpret escape sequences such as `\n`; just use
#   [double-quoted strings](language.html#double-quoted-string).
#
# See also [`print`](), [`echo`](), [`pprint`](), and [`repr`]().
fn printf {|template @value| }

# Print all arguments, joined by the `sep` option, and followed by a newline.
#
# Examples:
#
# ```elvish-transcript
# ~> echo Hello   elvish
# Hello elvish
# ~> echo "Hello   elvish"
# Hello   elvish
# ~> echo &sep=, lorem ipsum
# lorem,ipsum
# ```
#
# Notes: The `echo` builtin does not treat `-e` or `-n` specially. For instance,
# `echo -n` just prints `-n`. Use double-quoted strings to print special
# characters, and `print` to suppress the trailing newline.
#
# See also [`print`]().
#
# Etymology: Bourne sh.
fn echo {|&sep=' ' @value| }

# Pretty-print representations of Elvish values. Examples:
#
# ```elvish-transcript
# ~> pprint [foo bar]
# [
#  foo
#  bar
# ]
# ~> pprint [&k1=v1 &k2=v2]
# [
#  &k1=	v1
#  &k2=	v2
# ]
# ```
#
# The output format is subject to change.
#
# See also [`repr`]().
fn pprint {|@value| }

# Writes representation of `$value`s, separated by space and followed by a
# newline. Example:
#
# ```elvish-transcript
# ~> repr [foo 'lorem ipsum'] "aha\n"
# [foo 'lorem ipsum'] "aha\n"
# ```
#
# See also [`pprint`]().
#
# Etymology: [Python](https://docs.python.org/3/library/functions.html#repr).
fn repr {|@value| }

# Shows the value to the output, which is assumed to be a VT-100-compatible
# terminal.
#
# Currently, the only type of value that can be showed is exceptions, but this
# will likely expand in future.
#
# Example:
#
# ```elvish-transcript
# ~> var e = ?(fail lorem-ipsum)
# ~> show $e
# Exception: lorem-ipsum
#   [tty]:1:11-26: var e = ?(fail lorem-ipsum)
# ```
fn show {|e| }

# Passes byte input to output, and discards value inputs.
#
# Example:
#
# ```elvish-transcript
# ~> { put value; echo bytes } | only-bytes
# bytes
# ```
fn only-bytes { }

# Passes value input to output, and discards byte inputs.
#
# Example:
#
# ```elvish-transcript
# ~> { put value; echo bytes } | only-values
# ▶ value
# ```
fn only-values { }

# Reads bytes input into a single string, and put this string on structured
# stdout.
#
# Example:
#
# ```elvish-transcript
# ~> echo "a\nb" | slurp
# ▶ "a\nb\n"
# ```
#
# Etymology: Perl, as
# [`File::Slurp`](http://search.cpan.org/~uri/File-Slurp-9999.19/lib/File/Slurp.pm).
fn slurp { }

# Splits byte input into lines, and writes them to the value output. Value
# input is ignored.
#
# ```elvish-transcript
# ~> { echo a; echo b } | from-lines
# ▶ a
# ▶ b
# ~> { echo a; put b } | from-lines
# ▶ a
# ```
#
# See also [`from-terminated`](), [`read-upto`](), and [`to-lines`]().
fn from-lines { }

# Takes bytes stdin, parses it as JSON and puts the result on structured stdout.
# The input can contain multiple JSONs, and whitespace between them are ignored.
#
# Numbers in JSON are parsed as follows:
#
# -   Numbers without fractional parts are parsed as exact integers, with
#     support for arbitrary precision.
#
# -   Numbers with fractional parts (even if it's `.0`) are parsed as
#     [inexact](language.html#exactness) floating-point numbers, and the parsing
#     may fail if the number can't be represented.
#
# Examples:
#
# ```elvish-transcript
# ~> echo '"a"' | from-json
# ▶ a
# ~> echo '["lorem", "ipsum"]' | from-json
# ▶ [lorem ipsum]
# ~> echo '{"lorem": "ipsum"}' | from-json
# ▶ [&lorem=ipsum]
# ~> # multiple JSONs running together
#    echo '"a""b"["x"]' | from-json
# ▶ a
# ▶ b
# ▶ [x]
# ~> # multiple JSONs separated by newlines
#    echo '"a"
#    {"k": "v"}' | from-json
# ▶ a
# ▶ [&k=v]
# ~> echo '[42, 100000000000000000000, 42.0, 42.2]' | from-json
# ▶ [(num 42) (num 100000000000000000000) (num 42.0) (num 42.2)]
# ```
#
# See also [`to-json`]().
fn from-json { }

# Splits byte input into lines at each `$terminator` character, and writes
# them to the value output. If the byte input ends with `$terminator`, it is
# dropped. Value input is ignored.
#
# The `$terminator` must be a single ASCII character such as `"\x00"` (NUL).
#
# ```elvish-transcript
# ~> { echo a; echo b } | from-terminated "\x00"
# ▶ "a\nb\n"
# ~> print "a\x00b" | from-terminated "\x00"
# ▶ a
# ▶ b
# ~> print "a\x00b\x00" | from-terminated "\x00"
# ▶ a
# ▶ b
# ```
#
# See also [`from-lines`](), [`read-upto`](), and [`to-terminated`]().
fn from-terminated {|terminator| }

# Writes each input to a separate line in the byte output.
#
# ```elvish-transcript
# ~> put a b | to-lines
# a
# b
# ~> to-lines [a b]
# a
# b
# ```
#
# See also [`from-lines`]() and [`to-terminated`]().
fn to-lines {|inputs?| }

# Writes each input to the byte output with the specified terminator character.
#
# The `$terminator` must be a single ASCII character such as `"\x00"` (NUL).
# Using NUL is useful for feeding output to programs that expect NUL-terminated
# strings, such as `grep -z` or `xargs -0`.
#
# ```elvish-transcript
# ~> put a b | to-terminated "\x00" | slurp
# ▶ "a\x00b\x00"
# ~> to-terminated "\x00" [a b] | slurp
# ▶ "a\x00b\x00"
# ```
#
# See also [`from-terminated`]() and [`to-lines`]().
fn to-terminated {|terminator inputs?| }

# Takes structured stdin, convert it to JSON and puts the result on bytes stdout.
#
# ```elvish-transcript
# ~> put a | to-json
# "a"
# ~> put [lorem ipsum] | to-json
# ["lorem","ipsum"]
# ~> put [&lorem=ipsum] | to-json
# {"lorem":"ipsum"}
# ```
#
# See also [`from-json`]().
fn to-json { }
