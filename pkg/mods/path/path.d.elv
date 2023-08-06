# Compatibility alias for [`$os:dev-null`](). This variable will be formally
# deprecated and removed in future.
var dev-null

# Compatibility alias for [`$os:dev-tty`](). This variable will be formally
# deprecated and removed in future.
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

# Compatibility alias for [`os:eval-symlinks`](). This function will be formally
# deprecated and removed in future.
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

# Compatibility alias for [`os:is-dir`](). This function will be formally
# deprecated and removed in future.
fn is-dir {|&follow-symlink=$false path| }

# Compatibility alias for [`os:is-regular`](). This function will be formally
# deprecated and removed in future.
fn is-regular {|&follow-symlink=$false path| }

# Compatibility alias for [`os:temp-dir`](). This function will be formally
# deprecated and removed in future.
fn temp-dir {|&dir='' pattern?| }

# Compatibility alias for [`os:temp-file`](). This function will be formally
# deprecated and removed in future.
fn temp-file {|&dir='' pattern?| }
