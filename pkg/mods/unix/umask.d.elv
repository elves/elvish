#elvdoc:var umask
#
# The file mode creation mask. Its value is a string in Elvish octal
# representation; e.g. 0o027. This makes it possible to use it in any context
# that expects a `$number`.
#
# When assigning a new value a string is implicitly treated as an
# octal number. If that fails the usual rules for interpreting
# [numbers](./language.html#number) are used. The following are equivalent:
# `set unix:umask = 027` and `set unix:umask = 0o27`. You can also assign to it
# a `float64` data type that has no fractional component. The assigned value
# must be within the range [0 ... 0o777], otherwise the assignment will throw
# an exception.
#
# You can do a temporary assignment to affect a single command; e.g. `umask=077
# touch a_file`. After the command completes the old umask will be restored.
# **Warning**: Since the umask applies to the entire process, not individual
# threads, changing it temporarily in this manner is dangerous if you are doing
# anything in parallel, such as via the [`peach`](builtin.html#peach) command.
