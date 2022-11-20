#elvdoc:var completion:arg-completer
#
# A map containing argument completers.

#elvdoc:var completion:binding
#
# Keybinding for the completion mode.

#elvdoc:var completion:matcher
#
# A map mapping from context names to matcher functions. See the
# [Matcher](#matcher) section.

#elvdoc:fn complete-filename
#
# ```elvish
# edit:complete-filename $args...
# ```
#
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

#elvdoc:fn complex-candidate
#
# ```elvish
# edit:complex-candidate $stem &display='' &code-suffix=''
# ```
#
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

#elvdoc:fn match-prefix
#
# ```elvish
# edit:match-prefix $seed $inputs?
# ```
#
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

#elvdoc:fn match-subseq
#
# ```elvish
# edit:match-subseq $seed $inputs?
# ```
#
# For each input, outputs whether the input has $seed as a
# [subsequence](https://en.wikipedia.org/wiki/Subsequence). Uses the result of
# `to-string` for non-string inputs.

#elvdoc:fn match-substr
#
# ```elvish
# edit:match-substr $seed $inputs?
# ```
#
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

#elvdoc:fn completion:start
#
# ```elvish
# edit:completion:start
# ```
#
# Start the completion mode.

#elvdoc:fn completion:smart-start
#
# ```elvish
# edit:completion:smart-start
# ```
#
# Starts the completion mode. However, if all the candidates share a non-empty
# prefix and that prefix starts with the seed, inserts the prefix instead.

#elvdoc:fn completion:close
#
# ```elvish
# edit:completion:close
# ```
#
# Closes the completion mode UI.
