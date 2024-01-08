# The [file mode creation mask](https://en.wikipedia.org/wiki/Umask) for the
# process.
#
# Bits that are set in the mask get **cleared** in the actual permission of
# newly created files: for example, a value of `0o022` causes the group write
# and world write bits to be cleared when creating a new file.
#
# This variable has some special properties when read or assigned:
#
# -   When read, the value is a string in octal representation, like `0o027`.
#
# -   When assigned, both strings that parse as numbers and typed numbers may be
#     specified as the new value, as long as the value is within the range [0,
#     0o777].
#
#     As a special case for this variable, strings that don't start with `0b` or
#     `0x` are treated as octal instead of decimal: for example, `set unix:umask
#     = 27` is equivalent to `set unix:umask = 0o27` or `set unix:umask = (num
#     0o27)`, and **not** the same as `set unix:umask = (num 27)`.
#
# You can do a temporary assignment to affect a single command, like
# `{ tmp umask = 077; touch a_file }`, but beware that since umask applies to
# the whole process, any code that runs in parallel (such as via
# [`peach`]()) can also get affected.
var umask
