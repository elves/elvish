# A map containing argument completers.
var completion:arg-completer

# Keybinding for the completion mode.
var completion:binding

# A map mapping from context names to matcher functions. See the
# [Matcher](#matcher) section.
var completion:matcher

# Produces a list of filenames that are suitable for completing the last
# argument, ignoring all other arguments. The last argument is used in the
# following ways:
#
# - The directory determines which directory to complete.
#
# - If the base name starts with `.`, it completes all files whose names start
#   with `.` (hidden files); otherwise it completes filenames that don't start
#   with `.` (non-hidden files).
#
#   The rest of the base name is ignored; filtering is left to the
#   [matcher](#matcher).
#
# The outputs are [`edit:complex-candidate`]() objects, with styles determined
# by `$E:LSCOLOR`. Directories have a trailing `/` in the stem; non-directory
# files have a space as their code suffix.
#
# This function is the default handler for any commands without explicit
# handlers in `$edit:completion:arg-completer`. See [Argument
# Completer](#argument-completer).
#
# Example:
#
# ```elvish-transcript
# ~> ls -AR
# ~/tmp/example> ls -AR
# .ipsum .lorem bar    d      foo
#
# ./d:
# bar    foo
# ~> edit:complete-filename '' # non-hidden files in working directory
# ▶ (edit:complex-candidate bar &code-suffix=' ' &display=[^styled bar])
# ▶ (edit:complex-candidate d/ &code-suffix='' &display=[^styled (styled-segment d/ &fg-color=blue &bold)])
# ▶ (edit:complex-candidate foo &code-suffix=' ' &display=[^styled foo])
# ~> edit:complete-filename '.f' # hidden files in working directory
# ▶ (edit:complex-candidate .ipsum &code-suffix=' ' &display=[^styled .ipsum])
# ▶ (edit:complex-candidate .lorem &code-suffix=' ' &display=[^styled .lorem])
# ~> edit:complete-filename ./d/f # non-hidden files in ./d
# ▶ (edit:complex-candidate ./d/bar &code-suffix=' ' &display=[^styled ./d/bar])
# ▶ (edit:complex-candidate ./d/foo &code-suffix=' ' &display=[^styled ./d/foo])
# ```
fn complete-filename {|@args| }

#doc:added-in 0.21
#
# Like [`edit:complete-filename`](), but only generates directories.
fn complete-dirname {|@args| }

# Builds a complex candidate. This is mainly useful in [argument
# completers](#argument-completer).
#
# The `&display` option controls how the candidate is shown in the UI. It can
# be a string or a [styled](builtin.html#styled) text. If it is empty, `$stem`
# is used.
#
# The `&code-suffix` option affects how the candidate is inserted into the code
# when it is accepted. By default, a quoted version of `$stem` is inserted. If
# `$code-suffix` is non-empty, it is added to that text, and the suffix is not
# quoted.
fn complex-candidate {|stem &display='' &code-suffix=''| }

# For each input, outputs whether the input has $seed as a prefix. Uses the
# result of `to-string` for non-string inputs.
#
# Roughly equivalent to the following Elvish function, but more efficient:
#
# ```elvish
# use str
# fn match-prefix {|seed @input|
#   each {|x| str:has-prefix (to-string $x) $seed } $@input
# }
# ```
fn match-prefix {|seed inputs?| }

# For each input, outputs whether the input has $seed as a
# [subsequence](https://en.wikipedia.org/wiki/Subsequence). Uses the result of
# `to-string` for non-string inputs.
fn match-subseq {|seed inputs?| }

# For each input, outputs whether the input has $seed as a substring. Uses the
# result of `to-string` for non-string inputs.
#
# Roughly equivalent to the following Elvish function, but more efficient:
#
# ```elvish
# use str
# fn match-substr {|seed @input|
#   each {|x| str:has-contains (to-string $x) $seed } $@input
# }
# ```
fn match-substr {|seed inputs?| }

# Start the completion mode.
fn completion:start { }

# Starts the completion mode after accepting any pending autofix.
#
# If all the candidates share a non-empty prefix and that prefix starts with the
# seed, inserts the prefix instead.
fn completion:smart-start { }
