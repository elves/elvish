# OS-specific path list separator. Colon (`:`) on UNIX and semicolon (`;`) on
# Windows. This variable is read-only.
var list-separator

# OS-specific path separator. Forward slash (`/`) on UNIX and backslash (`\`)
# on Windows. This variable is read-only.
var separator

# Outputs `$path` converted to an absolute path.
#
# ```elvish-transcript
# ~> cd ~
# ~> path:abs bin
# ▶ /home/user/bin
# ```
fn abs {|path| }

# Outputs the last element of `$path`. This is analogous to the POSIX `basename` command. See the
# [Go documentation](https://pkg.go.dev/path/filepath#Base) for more details.
#
# ```elvish-transcript
# ~> path:base ~/bin
# ▶ bin
# ```
fn base {|path| }

# Outputs the shortest version of `$path` equivalent to `$path` by purely lexical processing. This
# is most useful for eliminating unnecessary relative path elements such as `.` and `..` without
# asking the OS to evaluate the path name. See the [Go
# documentation](https://pkg.go.dev/path/filepath#Clean) for more details.
#
# ```elvish-transcript
# ~> path:clean ./../bin
# ▶ ../bin
# ```
fn clean {|path| }

# Outputs all but the last element of `$path`, typically the path's enclosing directory. See the
# [Go documentation](https://pkg.go.dev/path/filepath#Dir) for more details. This is analogous to
# the POSIX `dirname` command.
#
# ```elvish-transcript
# ~> path:dir /a/b/c/something
# ▶ /a/b/c
# ```
fn dir {|path| }

# Outputs the file name extension used by `$path` (including the separating period). If there is no
# extension the empty string is output. See the [Go
# documentation](https://pkg.go.dev/path/filepath#Ext) for more details.
#
# ```elvish-transcript
# ~> path:ext hello.elv
# ▶ .elv
# ```
fn ext {|path| }

# Outputs `$true` if the path is an absolute path. Note that platforms like Windows have different
# rules than UNIX like platforms for what constitutes an absolute path. See the [Go
# documentation](https://pkg.go.dev/path/filepath#IsAbs) for more details.
#
# ```elvish-transcript
# ~> path:is-abs hello.elv
# ▶ false
# ~> path:is-abs /hello.elv
# ▶ true
# ```
fn is-abs {|path| }

# Outputs `$path` after resolving any symbolic links. If `$path` is relative the result will be
# relative to the current directory, unless one of the components is an absolute symbolic link.
# This function calls `path:clean` on the result before outputting it. This is analogous to the
# external `realpath` or `readlink` command found on many systems. See the [Go
# documentation](https://pkg.go.dev/path/filepath#EvalSymlinks) for more details.
#
# ```elvish-transcript
# ~> mkdir bin
# ~> ln -s bin sbin
# ~> path:eval-symlinks ./sbin/a_command
# ▶ bin/a_command
# ```
fn eval-symlinks {|path| }

# Joins any number of path elements into a single path, separating them with an
# [OS specific separator](#path:separator). Empty elements are ignored. The
# result is [cleaned](#path:clean). However, if the argument list is empty or
# all its elements are empty, Join returns an empty string. On Windows, the
# result will only be a UNC path if the first non-empty element is a UNC path.
#
# ```elvish-transcript
# ~> path:join home user bin
# ▶ home/user/bin
# ~> path:join $path:separator home user bin
# ▶ /home/user/bin
# ```
fn join {|@path-component| }

# Outputs `$true` if the path resolves to a directory. If the final element of the path is a
# symlink, even if it points to a directory, it still outputs `$false` since a symlink is not a
# directory. Setting option `&follow-symlink` to true will cause the last element of the path, if
# it is a symlink, to be resolved before doing the test.
#
# ```elvish-transcript
# ~> touch not-a-dir
# ~> path:is-dir not-a-dir
# ▶ false
# ~> path:is-dir /tmp
# ▶ true
# ```
#
# @cf path:is-regular
fn is-dir {|&follow-symlink=$false path| }

# Outputs `$true` if the path resolves to a regular file. If the final element of the path is a
# symlink, even if it points to a regular file, it still outputs `$false` since a symlink is not a
# regular file. Setting option `&follow-symlink` to true will cause the last element of the path,
# if it is a symlink, to be resolved before doing the test.
#
# **Note:** This isn't named `is-file` because a UNIX file may be a "bag of bytes" or may be a
# named pipe, device special file (e.g. `/dev/tty`), etc.
#
# ```elvish-transcript
# ~> touch not-a-dir
# ~> path:is-regular not-a-dir
# ▶ true
# ~> path:is-regular /tmp
# ▶ false
# ```
#
# @cf path:is-dir
fn is-regular {|&follow-symlink=$false path| }

# Creates a new directory and outputs its name.
#
# The &dir option determines where the directory will be created; if it is an
# empty string (the default), a system-dependent directory suitable for storing
# temporary files will be used. The `$pattern` argument determines the name of
# the directory, where the last star will be replaced by a random string; it
# defaults to `elvish-*`.
#
# It is the caller's responsibility to remove the directory if it is intended
# to be temporary.
#
# ```elvish-transcript
# ~> path:temp-dir
# ▶ /tmp/elvish-RANDOMSTR
# ~> path:temp-dir x-
# ▶ /tmp/x-RANDOMSTR
# ~> path:temp-dir 'x-*.y'
# ▶ /tmp/x-RANDOMSTR.y
# ~> path:temp-dir &dir=.
# ▶ elvish-RANDOMSTR
# ~> path:temp-dir &dir=/some/dir
# ▶ /some/dir/elvish-RANDOMSTR
# ```
fn temp-dir {|&dir='' pattern?| }

# Creates a new file and outputs a [file](language.html#file) object opened
# for reading and writing.
#
# The &dir option determines where the file will be created; if it is an
# empty string (the default), a system-dependent directory suitable for storing
# temporary files will be used. The `$pattern` argument determines the name of
# the file, where the last star will be replaced by a random string; it
# defaults to `elvish-*`.
#
# It is the caller's responsibility to close the file with
# [`file:close`](file.html#file:close). The caller should also remove the file
# if it is intended to be temporary (with `rm $f[name]`).
#
# ```elvish-transcript
# ~> var f = (path:temp-file)
# ~> put $f[name]
# ▶ /tmp/elvish-RANDOMSTR
# ~> echo hello > $f
# ~> cat $f[name]
# hello
# ~> var f = (path:temp-file x-)
# ~> put $f[name]
# ▶ /tmp/x-RANDOMSTR
# ~> var f = (path:temp-file 'x-*.y')
# ~> put $f[name]
# ▶ /tmp/x-RANDOMSTR.y
# ~> var f = (path:temp-file &dir=.)
# ~> put $f[name]
# ▶ elvish-RANDOMSTR
# ~> var f = (path:temp-file &dir=/some/dir)
# ~> put $f[name]
# ▶ /some/dir/elvish-RANDOMSTR
# ```
fn temp-file {|&dir='' pattern?| }
