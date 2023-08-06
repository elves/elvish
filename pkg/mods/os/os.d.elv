# OS-specific path to the "null" device: `/dev/null` on Unix and `NUL` on
# Windows.
var dev-null

# OS-specific path to the terminal device: `/dev/tty` on Unix and `CON` on
# Windows.
var dev-tty

#doc:show-unstable
#
# Reports whether an exception is caused by the fact that a file or directory
# already exists.
fn -is-exist {|exc| }

#doc:show-unstable
#
# Reports whether an exception is caused by the fact that a file or directory
# does not exist.
fn -is-not-exist {|exc| }

# Creates a new directory with the specified name and permission (before umask).
fn mkdir {|&perm=0o755 path| }

# Removes the file or empty directory at `path`.
#
# If the path does not exist, this command throws an exception that can be
# tested with [`os:-is-not-exist`]().
fn remove {|path| }

# Removes the named file or directory at `path` and, in the latter case, any
# children it contains. It removes everything it can, but returns the first
# error it encounters.
#
# If the path does not exist, this command returns silently without throwing an
# exception.
fn remove-all {|path| }

# Outputs `$path` after resolving any symbolic links. If `$path` is relative the result will be
# relative to the current directory, unless one of the components is an absolute symbolic link.
# This function calls `path:clean` on the result before outputting it. This is analogous to the
# external `realpath` or `readlink` command found on many systems. See the [Go
# documentation](https://pkg.go.dev/path/filepath#EvalSymlinks) for more details.
#
# ```elvish-transcript
# ~> mkdir bin
# ~> ln -s bin sbin
# ~> os:eval-symlinks ./sbin/a_command
# ▶ bin/a_command
# ```
fn eval-symlinks {|path| }

# Reports whether a file is known to exist at `path`.
#
# See [`follow-symlink`](#follow-symlink) for an explanation of the option.
fn exists {|&follow-symlink=$false path| }

# Reports whether a directory exists at the path.
#
# See [`follow-symlink`](#follow-symlink) for an explanation of the option.
#
# ```elvish-transcript
# ~> touch not-a-dir
# ~> os:is-dir not-a-dir
# ▶ false
# ~> os:is-dir /tmp
# ▶ true
# ```
#
# See also [`os:is-regular`]().
fn is-dir {|&follow-symlink=$false path| }

# Reports whether a regular file exists at the path.
#
# **Note:** Some other languages call this functionality something like
# `is-file`; that name is not chosen because the name "file" also includes
# things like directories and device files.
#
# ```elvish-transcript
# ~> touch not-a-dir
# ~> os:is-regular not-a-dir
# ▶ true
# ~> os:is-regular /tmp
# ▶ false
# ```
#
# See also [`os:is-dir`]().
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
# ~> os:temp-dir
# ▶ /tmp/elvish-RANDOMSTR
# ~> os:temp-dir x-
# ▶ /tmp/x-RANDOMSTR
# ~> os:temp-dir 'x-*.y'
# ▶ /tmp/x-RANDOMSTR.y
# ~> os:temp-dir &dir=.
# ▶ elvish-RANDOMSTR
# ~> os:temp-dir &dir=/some/dir
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
# ~> var f = (os:temp-file)
# ~> put $f[name]
# ▶ /tmp/elvish-RANDOMSTR
# ~> echo hello > $f
# ~> cat $f[name]
# hello
# ~> var f = (os:temp-file x-)
# ~> put $f[name]
# ▶ /tmp/x-RANDOMSTR
# ~> var f = (os:temp-file 'x-*.y')
# ~> put $f[name]
# ▶ /tmp/x-RANDOMSTR.y
# ~> var f = (os:temp-file &dir=.)
# ~> put $f[name]
# ▶ elvish-RANDOMSTR
# ~> var f = (os:temp-file &dir=/some/dir)
# ~> put $f[name]
# ▶ /some/dir/elvish-RANDOMSTR
# ```
fn temp-file {|&dir='' pattern?| }
