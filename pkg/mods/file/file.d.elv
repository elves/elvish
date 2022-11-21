#elvdoc:fn is-tty
#
# ```elvish
# file:is-tty $file
# ```
#
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

#elvdoc:fn open
#
# ```elvish
# file:open $filename
# ```
#
# Opens a file. Currently, `open` only supports opening a file for reading.
# File must be closed with `close` explicitly. Example:
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
# ~> close $f
# ```
#
# @cf file:close

#elvdoc:fn close
#
# ```elvish
# file:close $file
# ```
#
# Closes a file opened with `open`.
#
# @cf file:open

#elvdoc:fn pipe
#
# ```elvish
# file:pipe
# ```
#
# Create a new pipe that can be used in redirections. A pipe contains a read-end and write-end.
# Each pipe object is a [pseudo-map](language.html#pseudo-map) with fields `r` (the read-end [file
# object](./language.html#file)) and `w` (the write-end).
#
# When redirecting command input from a pipe with `<`, the read-end is used. When redirecting
# command output to a pipe with `>`, the write-end is used. Redirecting both input and output with
# `<>` to a pipe is not supported.
#
# Pipes have an OS-dependent buffer, so writing to a pipe without an active reader
# does not necessarily block. Pipes **must** be explicitly closed with `file:close`.
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
# @cf file:close

#elvdoc:fn truncate
#
# ```elvish
# file:truncate $filename $size
# ```
#
# changes the size of the named file. If the file is a symbolic link, it
# changes the size of the link's target. The size must be an integer between 0
# and 2^64-1.
