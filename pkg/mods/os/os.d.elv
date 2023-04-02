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

# Reports whether a file is known to exist at `path`.
fn exists {|path| }

# Output a [pseudo-map](language.html#pseudo-map) containing metadata for each
# filesystem path.
#
# If passed zero path names it does nothing; otherwise, it iterates over the
# list of path names and attempts to determine the characteristics of each
# path. Like the traditional Unix `stat` command an error processing a path
# name does not immediately terminate processing the list of path names. This
# command attempts to stat the remaining path names. This can result in a
# "multiple error" exception that documents each path that it could not stat
# while still outputting information about the other paths.
#
# If the `&follow-symlink` option is false, and the path refers to a symbolic
# link, then metadata about the symlink is output. If `&follow-symlink` is
# true then the metadata about the target of the symlink is output.
#
# Some keys may have a constant value depending on the platform. For example,
# Windows will always report `$false` for the `is-named-pipe` key since
# windows does not support named pipes. The following keys are populated on
# all platforms:
#
# - `abs-path`: The absolute path of the `path` value.
#
# - `is-char-device`: True if the path refers to a character device file.
#
# - `is-device`: True if the path refers to a device file.
#
# - `is-dir`: True if the path refers to a directory.
#
# - `is-named-pipe`: True if the path refers to a named pipe file.
#
# - `is-regular`: True if the path refers to a regular file.
#
# - `is-socket`: True if the path refers to a Unix domain socket file.
#
# - `is-symlink`: True if the path refers to a symbolic link file.
#
# - `m-time`: The modification time of the file.
#
# - `mode`: A number describing the "mode" of the file. This includes file
# permissions and other attributes that describe various aspects of the file
# (such as whether it is a regular file, directory, etc.). The value is
# defined by the [Go fs API](https://pkg.go.dev/io/fs#FileMode).
#
# - `path`: The original path passed to the command.
#
# - `perms`: A number describing the permissions and set-uid, set-gid, and
# sticky attributes of the file when interpreted as a bit pattern. The meaning
# of this value depends on the platform. Note that on Windows only the user
# write permission bit (0o200) is meangingful.
#
# - `size`: The size of the file in bytes.
#
# - `symbolic-mode`: A string representation of the `mode` value.
#
# - `symbolic-perms`: A string representation of the `perms` value.
#
# These keys are always present in the pseudo-map but might not be initialized
# (thus having the "zero value") depending on the platform:
#
# - `a-time`: The access time of the file. Meaningful on Unix and Windows.
#
# - `b-time`: The birth (i.e., creation) time of the file. Meaningful on
# FreeBSD, NetBSD, Darwin (macOS), and Windows.
#
# - `block-count`: The size of the file in blocks of `block-size`. Meaningful
# on Unix.
#
# - `block-size`: The block size for I/O. Meaningful on Unix.
#
# - `c-time`: The status change time of the file. Status changes include
# events such as changing the owner or permissions of the file. Meaningful on
# Unix.
#
# - `device`: The device ID of the filesystem containing file. Meaningful on Unix.
#
# - `gid`: The group ID that owns the file. Meaningful on Unix.
#
# - `group`: The group name for the `gid` that owns the file. Meaningful on Unix.
#
# - `inode`: The inode number of the file. Meaningful on Unix.
#
# - `num-links`: The number of hard links to the file. Meaningful on Unix.
#
# - `owner`: The user name for the `uid` that owns the file. Meaningful on Unix.
#
# - `raw-device`: The raw device ID if the file is a device node (rather than
# a directory, regular file, or symlink). Meaningful on Unix.
#
# - `uid`: The user ID that owns the file. Meaningful on Unix.
#
# Example:
#
# ```elvish-transcript
# ~> nop > f
# ~> ls -l f
# -rw-r----- 1 krader staff 0 Aug  4 21:27 f
# ~> pprint (os:stat f)
# [
#  &a-time=       <time{2023-08-04 21:27:17.777839668 -0700 PDT}>
#  &abs-path=     /Users/krader/projects/3rd-party/elvish/f
#  &b-time=       <time{2023-08-04 21:27:17.777839668 -0700 PDT}>
#  &block-count=  (num 0)
#  &block-size=   (num 4096)
#  &c-time=       <time{2023-08-04 21:27:17.777880376 -0700 PDT}>
#  &device=       (num 16777229)
#  &gid=  (num 20)
#  &group=        staff
#  &inode=        (num 77318437)
#  &is-char-device=       $false
#  &is-device=    $false
#  &is-dir=       $false
#  &is-named-pipe=        $false
#  &is-regular=   $true
#  &is-socket=    $false
#  &is-symlink=   $false
#  &m-time=       <time{2023-08-04 21:27:17.777839668 -0700 PDT}>
#  &mode= (num 416)
#  &num-links=    (num 1)
#  &owner=        krader
#  &path= f
#  &perms=        (num 416)
#  &raw-device=   (num 0)
#  &size= (num 0)
#  &symbolic-mode=        -rw-r-----
#  &symbolic-perms=       -rw-r-----
#  &uid=  (num 501)
# ]
# ```
#
# See also [`path:is-dir`](), [`path:is-regular`]().
fn stat {|&follow-symlink=$false path...| }
