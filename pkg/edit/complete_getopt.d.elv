# Produces completions according to a specification of accepted command-line
# options (both short and long options are handled), positional handler
# functions for each command position, and the current arguments in the command
# line. The arguments are as follows:
#
# * `$args` is an array containing the current arguments in the command line
#   (without the command itself). These are the arguments as passed to the
#   [Argument Completer](#argument-completer) function.
#
# * `$opt-specs` is an array of maps, each one containing the definition of
#   one possible command-line option. Matching options will be provided as
#   completions when the last element of `$args` starts with a dash, but not
#   otherwise. Each map can contain the following keys (at least one of `short`
#   or `long` needs to be specified):
#
#   - `short` contains the one-letter short option, if any, without the dash.
#
#   - `long` contains the long option name, if any, without the initial two
#     dashes.
#
#   - `arg-optional`, if set to `$true`, specifies that the option receives an
#     optional argument.
#
#   - `arg-required`, if set to `$true`, specifies that the option receives a
#     mandatory argument. Only one of `arg-optional` or `arg-required` can be
#     set to `$true`.
#
#   - `desc` can be set to a human-readable description of the option which
#     will be displayed in the completion menu.
#
#   - `completer` can be set to a function to generate possible completions for
#     the option argument. The function receives as argument the element at
#     that position and return zero or more candidates.
#
# * `$arg-handlers` is an array of functions, each one returning the possible
#   completions for that position in the arguments. Each function receives
#   as argument the last element of `$args`, and should return zero or more
#   possible values for the completions at that point. The returned values can
#   be plain strings or the output of `edit:complex-candidate`. If the last
#   element of the list is the string `...`, then the last handler is reused
#   for all following arguments.
#
# Example:
#
# ```elvish-transcript
# ~> fn complete {|@args|
#      opt-specs = [ [&short=a &long=all &desc="Show all"]
#                    [&short=n &desc="Set name" &arg-required=$true
#                     &completer= {|_| put name1 name2 }] ]
#      arg-handlers = [ {|_| put first1 first2 }
#                       {|_| put second1 second2 } ... ]
#      edit:complete-getopt $args $opt-specs $arg-handlers
#    }
# ~> complete ''
# ▶ first1
# ▶ first2
# ~> complete '-'
# ▶ (edit:complex-candidate -a &display='-a (Show all)')
# ▶ (edit:complex-candidate --all &display='--all (Show all)')
# ▶ (edit:complex-candidate -n &display='-n (Set name)')
# ~> complete -n ''
# ▶ name1
# ▶ name2
# ~> complete -a ''
# ▶ first1
# ▶ first2
# ~> complete arg1 ''
# ▶ second1
# ▶ second2
# ~> complete arg1 arg2 ''
# ▶ second1
# ▶ second2
# ```
#
# See also [`flag:parse-getopt`]().
fn complete-getopt {|args opt-specs arg-handlers| }
