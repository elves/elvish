# OS-specific path to the "null" device (`/dev/null` on Unix and `NUL` on
# Windows).
var dev-null

# OS-specific path to the terminal device (`/dev/tty` on Unix and `CON` on
# Windows).
var dev-tty

# OS-specific path list separator. Colon (`:`) on Unix and semicolon (`;`) on
# Windows. This variable is read-only.
var list-separator

# OS-specific path separator. Forward slash (`/`) on Unix and backslash (`\`)
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
# rules than Unix like platforms for what constitutes an absolute path. See the [Go
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
# [OS specific separator](#$path:separator). Empty elements are ignored. The
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
# See also [`path:is-regular`]().
fn is-dir {|&follow-symlink=$false path| }

# Outputs `$true` if the path resolves to a regular file. If the final element of the path is a
# symlink, even if it points to a regular file, it still outputs `$false` since a symlink is not a
# regular file. Setting option `&follow-symlink` to true will cause the last element of the path,
# if it is a symlink, to be resolved before doing the test.
#
# **Note:** This isn't named `is-file` because a Unix file may be a "bag of bytes" or may be a
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
# See also [`path:is-dir`]().
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

# Output a pseudo-map containing metadata about one or more paths.
#
# If passed zero path names it does nothing; otherwise, it iterates over the
# list of path names and attempts to determine the characteristics of each
# path. It outputs a [psuedo-map](language.html#pseudo-map) that exposes the
# metadata about each path.
#
# Like the traditional Unix `stat` command an error processing a path name
# does not immediately terminate processing the list of path names. This
# command attempts to stat the remaining path names. This can result in a
# "multiple error" exception that documents each path that could not stat'd
# while still outputting information about some paths.
#
# If the `&follow-symlink` option is false and the path refers to a symbolic
# link then metadata about the symlink is output. If `&follow-symlink` is
# true then the metadata about the target of the symlink is output.
#
# The following keys are available in the pseudo-map. Whether the key has a
# meaningful value depends on the platform.
#
# - `path`: The original path passed to the command. Available on all
# platforms.
#
# - `abs-path`: The absolute path of the `path` value. Available on all
# platforms.
#
# - `is-dir`: True if the path refers to a directory; otherwise, false.
# Available on all platforms.
#
# - `size`: The size of the file in bytes. Available on all platforms.
#
# - `mode`: A numeric value describing the permissions and other attributes
# of the file when interpreted as a bit pattern. The meaning of this value
# depends on the platform. Available on all platforms.
#
# - `symbolic-mode`: A string representation of the `mode` value. Available on
# all platforms.
#
# - `m-time`: The modification time of the file. Available on all platforms.
#
# - `a-time`: The access time of the file. Available on Unix and Windows.
#
# - `b-time`: The birth (i.e., creation) time of the file. Available on
# Darwin (macOS) and Windows.
#
# - `c-time`: The status change time of the file. Status changes include
# events such as changing the owner or permissions of the file. Available on
# Unix.
#
# - `owner`: The user name for the `uid` that owns the file. Available on Unix.
#
# - `group`: The group name for the `gid` that owns the file. Available on Unix.
#
# - `uid`: The user ID that owns the file. Available on Unix.
#
# - `gid`: The group ID that owns the file. Available on Unix.
#
# - `num-links`: The number of hard links to the file. Available on Unix.
#
# - `inode`: The inode number of the file. Available on Unix.
#
# - `device`: The device ID of the filesystem containing file. Available on Unix.
#
# - `raw-device`: The raw device ID if the file is a device node (rather than
# a directory, regular file, or symlink). Available on Unix.
#
# - `block-count`: The size of the file in blocks of `block-size`. Available on Unix.
#
# - `block-size`: The block size for I/O. Available on Unix.
#
# Example:
#
# ```elvish-transcript
# ~> pwd
# /tmp
# ~> nop > f
# ~> ls -l f
# -rw-r----- 1 krader staff 0 Apr  1 19:53 f
# ~> pprint (path:stat f)
# [
#  &path= f
#  &abs-path=     /tmp/f
#  &is-dir=       $false
#  &size= (num 0)
#  &mode= (num 416)
#  &symbolic-mode=        -rw-r-----
#  &m-time=       <unknown 2023-04-04 13:48:33.672098599 -0700 PDT>
#  &a-time=       <unknown 2023-04-04 13:48:33.672098599 -0700 PDT>
#  &b-time=       <unknown 2023-04-04 13:48:33.672098599 -0700 PDT>
#  &c-time=       <unknown 2023-04-04 13:48:33.672098599 -0700 PDT>
#  &owner=        krader
#  &group=        staff
#  &uid=  (num 501)
#  &gid=  (num 20)
#  &num-links=    (num 1)
#  &inode=        (num 55886574)
#  &device=       (num 16777229)
#  &raw-device=   (num 0)
#  &block-size=   (num 4096)
#  &block-count=  (num 0)
# ]
# ```
fn stat {|&follow-symlink=$false path...| }
