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

# Change the mode of `$path`.
#
# The `$mode` argument must be an integer. It is used as the absolute mode
# value to assign to `$path`. If `$mode` is a string rather than a
# [`num`](language.html#number) it can have a radix prefix (e.g., `0o` to
# indicate the following digits are octal). If there is not an explicit radix
# prefix octal is assumed. If that implicit radix conversion fails then the
# value is assumed to be radix 10. This is different from nearly all other
# Elvish string to number conversions which assume a string without an
# explicit radix are decimal.
#
# On Unix systems the permission bits (0o777, representing the owner, group,
# and other permission groups), SetUid (0o4000), SetGid (0o2000), and Sticky
# (0o1000) bits are valid. The permission bits are:
#
# - 0o400 Allow reads by the owner.
# - 0o200 Allow writes by the owner.
# - 0o100 Allow execution by the owner if a file, else search if a directory.
# - 0o040 Allow reads by group members.
# - 0o020 Allow writes by group members.
# - 0o010 Allow execution by group members if a file, else search if a directory.
# - 0o004 Allow reads by others.
# - 0o002 Allow writes by others.
# - 0o001 Allow execution by others if a file, else search if a directory.
#
# On Windows only the user write bit (0o200) is valid (the other bits are
# ignored). If that bit is set the file mode allows writing; otherwise, the
# file is read-only.
#
# ```elvish-transcript
# ~> echo > f
# ~> ls -l f
# -rw-r----- 1 user group 1 Mar 29 20:28 f
# ~> os:chmod 0o2123 f
# ~> ls -l f # Note the `S` for the SetGid mode
# ---x-wS-wx 1 user group 1 Mar 29 20:28 f*
# ```
fn chmod {|mode path| }

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

# Describes the file at `path` by writing a map with the following fields:
#
# - `name`: The base name of the file.
#
# - `size`: Length in bytes for regular files; system-dependent for others.
#
# - `type`: One of `regular`, `dir`, `symlink`, `named-pipe`, `socket`,
#   `device`, `char-device` and `irregular`.
#
# - `perm`: Permission bits of the file following Unix's convention.
#
#   See [numeric notation of Unix
#   permission](https://en.wikipedia.org/wiki/File-system_permissions#Numeric_notation)
#   for a description of the convention, but note that Elvish prints this number
#   in decimal; use [`printf`](builtin.html#printf) with `%o` to print it as an octal number.
#
# - `special-modes`: A list containing one or more of `setuid`, `setgid` and
#   `sticky` to indicate the presence of any special mode.
#
# - `sys`: System-dependent information:
#
#   - On Unix, a map that corresponds 1:1 to the `stat_t` struct, except that
#     timestamp fields are not exposed yet.
#
#   - On Windows, a map that currently contains just one field,
#     `file-attributes`, which is a list describing which [file attribute
#     fields](https://learn.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants)
#     are set. For example, if `FILE_ATTRIBUTE_READONLY` is set, the list
#     contains `readonly`.
#
# See [`follow-symlink`](#follow-symlink) for an explanation of the option.
#
# Examples:
#
# ```elvish-transcript
# ~> echo content > regular
# ~> os:stat regular
# ▶ [&name=regular &perm=(num 420) &size=(num 8) &special-modes=[] &sys=[&...] &type=regular]
# ~> mkdir dir
# ~> os:stat dir
# ▶ [&name=dir &perm=(num 493) &size=(num 96) &special-modes=[] &sys=[&...] &type=dir]
# ~> ln -s dir symlink
# ~> os:stat symlink
# ▶ [&name=symlink &perm=(num 493) &size=(num 3) &special-modes=[] &sys=[&...] &type=symlink]
# ~> os:stat &follow-symlink symlink
# ▶ [&name=symlink &perm=(num 493) &size=(num 96) &special-modes=[] &sys=[&...] &type=dir]
# ```
fn stat {|&follow-symlink=$false path| }

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
