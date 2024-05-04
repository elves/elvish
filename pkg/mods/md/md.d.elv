#//each:eval use md

#doc:added-in 0.21
# Renders `$markdown` in the terminal.
#
# The `&width` option specifies the width to wrap the output to. If it is 0 (the
# default) or negative, `show` queries the terminal width of the standard output
# and use it as the width, falling back to 80 if the query fails (for example
# when the standard output is not a terminal).
#
# Examples:
#
# ```elvish-transcript
# ~> md:show "#h1 heading\n- List\n- Item"
# #h1 heading
#
# • List
#
# • Item
# ```
#
# See also [`doc:show`]().
fn show {|&width=0| markdown}
