# A map containing argument completers.
var completion:arg-completer

# Keybinding for the completion mode.
var completion:binding

# A map mapping from context names to matcher functions. See the
# [Matcher](#matcher) section.
var completion:matcher

# Produces a list of filenames found in the directory of the last argument. All
# other arguments are ignored. If the last argument does not contain a path
# (either absolute or relative to the current directory), then the current
# directory is used. Relevant files are output as `edit:complex-candidate`
# objects.
#
# This function is the default handler for any commands without
# explicit handlers in `$edit:completion:arg-completer`. See [Argument
# Completer](#argument-completer).
#
# Example:
#
# ```elvish-transcript
# ~> edit:complete-filename ''
# ▶ (edit:complex-candidate Applications &code-suffix=/ &style='01;34')
# ▶ (edit:complex-candidate Books &code-suffix=/ &style='01;34')
# ▶ (edit:complex-candidate Desktop &code-suffix=/ &style='01;34')
# ▶ (edit:complex-candidate Docsafe &code-suffix=/ &style='01;34')
# ▶ (edit:complex-candidate Documents &code-suffix=/ &style='01;34')
# ...
# ~> edit:complete-filename .elvish/
# ▶ (edit:complex-candidate .elvish/aliases &code-suffix=/ &style='01;34')
# ▶ (edit:complex-candidate .elvish/db &code-suffix=' ' &style='')
# ▶ (edit:complex-candidate .elvish/epm-installed &code-suffix=' ' &style='')
# ▶ (edit:complex-candidate .elvish/lib &code-suffix=/ &style='01;34')
# ▶ (edit:complex-candidate .elvish/rc.elv &code-suffix=' ' &style='')
# ```
fn complete-filename {|@args| }

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

# Closes the completion mode UI.
fn completion:close { }
