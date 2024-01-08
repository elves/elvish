# Parses flags from `$args` according to the signature of the
# `$fn`, using the [Go convention](#go-convention), and calls `$fn`.
#
# The `$fn` must be a user-defined function (i.e. not a builtin
# function or external command). Each option corresponds to a flag; see
# [`flag:parse`]() for how the default value affects the behavior of flags.
# After parsing, the non-flag arguments are used as function arguments.
#
# Example:
#
# ```elvish-transcript
# ~> use flag
# ~> fn f {|&verbose=$false &port=(num 8000) name| put $verbose $port $name }
# ~> flag:call $f~ [-verbose -port 80 a.c]
# ▶ $true
# ▶ (num 80)
# ▶ a.c
# ```
#
# This function is most useful when creating an Elvish script that accepts
# command-line arguments. For example, if a script `a.elv` contains the
# following code:
#
# ```elvish
# use flag
# fn main { |&verbose=$false &port=(num 8000) name|
#   ...
# }
# flag:call $main~ $args
# ```
#
# The script can be used as follows:
#
# ```elvish-transcript
# ~> elvish a.elv -verbose -port 80 foo
# ...
# ```
#
# See also [`flag:parse`]().
fn call {|fn args| }

# Parses flags from `$args` according to the `$specs`, using the [Go
# convention](#go-convention).
#
# The `$args` must be a list of strings containing the command-line arguments
# to parse.
#
# The `$specs` must be a list of flag specs:
#
# ```elvish
# [
#   [flag default-value 'description of the flag']
#   ...
# ]
# ```
#
# Each flag spec consists of the name of the flag (without the leading `-`),
# its default value, and a description. The default value influences the how
# the flag gets converted from string:
#
# -   If it is boolean, the flag is a boolean flag (see [Go
#     convention](#go-convention) for implications). Flag values `0`, `f`, `F`,
#     `false`, `False` and `FALSE` are converted to `$false`, and `1`, `t`,
#     `T`, `true`, `True` and `TRUE` to `$true`. Other values are invalid.
#
# -   If it is a string, no conversion is done.
#
# -   If it is a [typed number](language.html#number), the flag value is
#     converted using [`num`]().
#
# -   If it is a list, the flag value is split at `,` (equivalent to `{|s| put
#     [(str:split , $s)] }`).
#
# -   If it is none of the above, an exception is thrown.
#
# On success, this command outputs two values: a map containing the value of
# flags defined in `$specs` (whether they appear in `$args` or not), and a list
# containing non-flag arguments.
#
# Example:
#
# ```elvish-transcript
# ~> flag:parse [-v -times 10 foo] [
#      [v $false 'Verbose']
#      [times (num 1) 'How many times']
#    ]
# ▶ [&v=$true &times=(num 10)]
# ▶ [foo]
# ~> flag:parse [] [
#      [v $false 'Verbose']
#      [times (num 1) 'How many times']
#    ]
# ▶ [&v=$false &times=(num 1)]
# ▶ []
# ```
#
# See also [`flag:call`]() and [`flag:parse-getopt`]().
fn parse {|args specs| }

# Parses flags from `$args` according to the `$specs`, using the [getopt
# convention](#getopt-convention) (see there for the semantics of the options),
# and outputs the result.
#
# The `$args` must be a list of strings containing the command-line arguments
# to parse.
#
# The `$specs` must be a list of flag specs:
#
# ```elvish
# [
#   [&short=f &long=flag &arg-optional=$false &arg-required=$false]
#   ...
# ]
# ```
#
# Each flag spec can contain the following:
#
# -   The short and long form of the flag, without the leading `-` or `--`. The
#     short form, if non-empty, must be one character. At least one of `&short`
#     and `&long` must be non-empty.
#
# -   Whether the flag takes an optional argument or a required argument. At
#     most one of `&arg-optional` and `&arg-required` may be true.
#
# It is not an error for a flag spec to contain more keys.
#
# On success, this command outputs two values: a list describing all flags
# parsed from `$args`, and a list containing non-flag arguments. The former
# list looks like:
#
# ```elvish
# [
#   [&spec=... &arg=value &long=$false]
#   ...
# ]
# ```
#
# Each entry contains the original spec for the flag, its argument, and whether
# the flag appeared in its long form.
#
# Example (some output reformatted for readability):
#
# ```elvish-transcript
# ~> var specs = [
#      [&short=v &long=verbose]
#      [&short=p &long=port &arg-required]
#    ]
# ~> flag:parse-getopt [-v -p 80 foo] $specs
# ▶ [[&spec=[&short=v &long=verbose] &long=$false &arg='']
#    [&spec=[&arg-required=$true &short=p &long=port] &long=$false &arg=80]]
# ▶ [foo]
# ~> flag:parse-getopt [--verbose] $specs
# ▶ [[&spec=[&short=v &long=verbose] &long=$true &arg='']]
# ▶ []
# ~> flag:parse-getopt [-v] [[&short=v &extra-info=foo]] # extra key in spec
# ▶ [[&spec=[&extra-info=foo &short=v] &long=$false &arg='']]
# ▶ []
# ```
#
# See also [`flag:parse`]() and [`edit:complete-getopt`]().
fn parse-getopt {|args specs &stop-after-double-dash=$true &stop-before-non-flag=$false &long-only=$false| }
