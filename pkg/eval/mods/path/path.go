// Package path provides functions for manipulating filesystem path names.
package path

import (
	"os"
	"path/filepath"

	"src.elv.sh/pkg/eval"
)

// Ns is the namespace for the re: module.
var Ns = eval.NsBuilder{}.AddGoFns("path:", fns).Ns()

var fns = map[string]interface{}{
	"abs":           filepath.Abs,
	"base":          filepath.Base,
	"clean":         filepath.Clean,
	"dir":           filepath.Dir,
	"ext":           filepath.Ext,
	"eval-symlinks": filepath.EvalSymlinks,
	"is-abs":        filepath.IsAbs,
	"is-dir":        isDir,
	"is-regular":    isRegular,
}

//elvdoc:fn abs
//
// ```elvish
// path:abs $path
// ```
//
// Outputs `$path` converted to an absolute path.
//
// ```elvish-transcript
// ~> cd ~
// ~> path:abs bin
// ▶ /home/user/bin
// ```

//elvdoc:fn base
//
// ```elvish
// path:base $path
// ```
//
// Outputs the last element of `$path`. This is analogous to the POSIX `basename` command. See the
// [Go documentation](https://pkg.go.dev/path/filepath#Base) for more details.
//
// ```elvish-transcript
// ~> path:base ~/bin
// ▶ bin
// ```

//elvdoc:fn clean
//
// ```elvish
// path:clean $path
// ```
//
// Outputs the shortest version of `$path` equivalent to `$path` by purely lexical processing. This
// is most useful for eliminating unnecessary relative path elements such as `.` and `..` without
// asking the OS to evaluate the path name. See the [Go
// documentation](https://pkg.go.dev/path/filepath#Clean) for more details.
//
// ```elvish-transcript
// ~> path:clean ./../bin
// ▶ ../bin
// ```

//elvdoc:fn dir
//
// ```elvish
// path:dir $path
// ```
//
// Outputs all but the last element of `$path`, typically the path's enclosing directory. See the
// [Go documentation](https://pkg.go.dev/path/filepath#Dir) for more details. This is analogous to
// the POSIX `dirname` command.
//
// ```elvish-transcript
// ~> path:dir /a/b/c/something
// ▶ /a/b/c
// ```

//elvdoc:fn ext
//
// ```elvish
// ext $path
// ```
//
// Outputs the file name extension used by `$path` (including the separating period). If there is no
// extension the empty string is output. See the [Go
// documentation](https://pkg.go.dev/path/filepath#Ext) for more details.
//
// ```elvish-transcript
// ~> path:ext hello.elv
// ▶ .elv
// ```

//elvdoc:fn is-abs
//
// ```elvish
// is-abs $path
// ```
//
// Outputs `$true` if the path is an abolute path. Note that platforms like Windows have different
// rules than UNIX like platforms for what constitutes an absolute path. See the [Go
// documentation](https://pkg.go.dev/path/filepath#IsAbs) for more details.
//
// ```elvish-transcript
// ~> path:is-abs hello.elv
// ▶ false
// ~> path:is-abs /hello.elv
// ▶ true
// ```

//elvdoc:fn eval-symlinks
//
// ```elvish-transcript
// ~> mkdir bin
// ~> ln -s bin sbin
// ~> path:eval-symlinks ./sbin/a_command
// ▶ bin/a_command
// ```
//
// Outputs `$path` after resolving any symbolic links. If `$path` is relative the result will be
// relative to the current directory, unless one of the components is an absolute symbolic link.
// This function calls `path:clean` on the result before outputing it. This is analogous to the
// external `realpath` or `readlink` command found on many systems. See the [Go
// documentation](https://pkg.go.dev/path/filepath#EvalSymlinks) for more details.

//elvdoc:fn is-dir
//
// ```elvish
// is-dir $path
// ```
//
// Outputs `$true` if the path resolves to a directory. If the final element of the path is a
// symlink, even if it points to a directory, it still outputs `$false` since a symlink is not a
// directory. Use [`eval-symlinks`](#patheval-symlinks) on the path first if you don't care if the
// final element is a symlink.
//
// ```elvish-transcript
// ~> touch not-a-dir
// ~> path:is-dir not-a-dir
// ▶ false
// ~> path:is-dir /tmp
// ▶ true
// ```

func isDir(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode().IsDir()
}

//elvdoc:fn is-regular
//
// ```elvish
// is-regular $path
// ```
//
// Outputs `$true` if the path resolves to a regular file. If the final element of the path is a
// symlink, even if it points to a regular file, it still outputs `$false` since a symlink is not a
// regular file. Use [`eval-symlinks`](#patheval-symlinks) on the path first if you don't care if
// the final element is a symlink.
//
// ```elvish-transcript
// ~> touch not-a-dir
// ~> path:is-regular not-a-dir
// ▶ true
// ~> path:is-dir /tmp
// ▶ false
// ```

func isRegular(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode().IsRegular()
}
