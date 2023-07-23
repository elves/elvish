# Outputs whether `$file` is a terminal device.
#
# The `$file` can be a file object or a number. If it's a number, it's
# interpreted as the number of an [IO port](language.html#io-ports).
#
# ```elvish-transcript
# ~> var f = (file:open /dev/tty)
# ~> file:is-tty $f
# ▶ $true
# ~> file:close $f
# ~> var f = (file:open /dev/null)
# ~> file:is-tty $f
# ▶ $false
# ~> file:close $f
# ~> var p = (file:pipe)
# ~> file:is-tty $p[r]
# ▶ $false
# ~> file:is-tty $p[w]
# ▶ $false
# ~> file:close $p[r]
# ~> file:close $p[w]
# ~> file:is-tty 0
# ▶ $true
# ~> file:is-tty 1
# ▶ $true
# ~> file:is-tty 2
# ▶ $true
# ~> file:is-tty 0 < /dev/null
# ▶ $false
# ~> file:is-tty 0 < /dev/tty
# ▶ $true
# ```
fn is-tty {|file| }

# Opens a file for input. The file must be closed with [`file:close`]() when no
# longer needed.
#
# Example:
#
# ```elvish-transcript
# ~> cat a.txt
# This is
# a file.
# ~> use file
# ~> var f = (file:open a.txt)
# ~> cat < $f
# This is
# a file.
# ~> file:close $f
# ```
#
# See also [`file:open-output`]() and [`file:close`]().
fn open {|filename| }

# Opens a file for output. The file must be closed with [`file:close`]() when no
# longer needed.
#
# If `&also-input` is true, the file may also be used for input.
#
# The `&if-not-exists` option can be either `create` or `error`.
#
# The `&if-exists` option can be either `truncate` (removing all data), `append`
# (appending to the end), `update` (updating in place) or `error`. The `error`
# value may only be used with `&if-not-exists=create`.
#
# The `&create-perm` option specifies what permission to create the file with if
# the file doesn't exist and `&if-not-exists=create`. It must be an integer
# within [0, 0o777]. On Unix, the actual file permission is subject to filtering
# by [`$unix:umask`]().
#
# Example:
#
# ```elvish-transcript
# ~> use file
# ~> var f = (file:open-output new)
# ~> echo content > $f
# ~> file:close $f
# ~> cat new
# content
# ```
#
# See also [`file:open`]() and [`file:close`]().
fn open-output {|filename &also-input=$false &if-not-exists=create &if-exists=truncate &create-perm=(num 0o644)| }

# Closes a file opened with `open`.
#
# See also [`file:open`]().
fn close {|file| }

# Creates a new pipe that can be used in redirections. Outputs a map with two
# fields: `r` contains the read-end of the pipe, and `w` contains the write-end.
# Both are [file object](language.html#file))
#
# When redirecting command input from a pipe with `<`, the read-end is used. When redirecting
# command output to a pipe with `>`, the write-end is used. Redirecting both input and output with
# `<>` to a pipe is not supported.
#
# Pipes have an OS-dependent buffer, so writing to a pipe without an active
# reader does not necessarily block. Both ends of the pipes must be explicitly
# closed with `file:close`.
#
# Putting values into pipes will cause those values to be discarded.
#
# Examples (assuming the pipe has a large enough buffer):
#
# ```elvish-transcript
# ~> var p = (file:pipe)
# ~> echo 'lorem ipsum' > $p
# ~> head -n1 < $p
# lorem ipsum
# ~> put 'lorem ipsum' > $p
# ~> file:close $p[w] # close the write-end
# ~> head -n1 < $p # blocks unless the write-end is closed
# ~> file:close $p[r] # close the read-end
# ```
#
# See also [`file:close`]().
fn pipe { }

# Sets the offset for the next read or write operation on `$file`.
#
# The `&whence` option specifies what the offset is relative to, and can be
# `start`, `current` or `end`.
#
# The behavior of `seek` on a file opened using [`file:open-output`]() with
# `&if-exists=append` is unspecified.
#
# Example:
#
# ```elvish-transcript
# ~> print 0123456789 > file
# ~> var f = (file:open file)
# ~> read-bytes 3 < $f
# ▶ 012
# ~> file:seek $f 0
# ~> read-bytes 3 < $f
# ▶ 012
# ~> file:seek $f 2 &whence=current
# ~> read-bytes 3 < $f
# ▶ 567
# ~> file:seek $f -3 &whence=end
# ~> read-bytes 3 < $f
# ▶ 789
# ~> file:close $f
# ```
fn seek {|file offset &whence=start| }

# Outputs the current offset within `$file`, relative to the start of the file.
#
# Example:
#
# ```elvish-transcript
# ~> print 0123456789 > file
# ~> var f = (file:open file)
# ~> read-bytes 3 < $f
# ▶ 012
# ~> file:tell $f
# ▶ (num 3)
# ~> file:close $f
# ```
fn tell {|file| }

# changes the size of the named file. If the file is a symbolic link, it
# changes the size of the link's target. The size must be an integer between 0
# and 2^64-1.
fn truncate {|filename size| }
