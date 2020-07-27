// Package path exposes functionality from Go's path package as an Elvish
// module.
package path

import (
	"path/filepath"

	"github.com/elves/elvish/pkg/eval"
)

//elvdoc:fn abs
//
// ```elvish
// path:abs $path
// ```
//
// Outputs the absolute path of `$path`.
//
// ```elvish-transcript
// ~> path:abs ~/bin
// ▶ /home/user/bin
// ```

//elvdoc:fn base
//
// ```elvish
// path:base $path
// ```
//
// Outputs the base path of `$path`.
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
// Outputs `$path` with leading `./` removed.
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
// Outputs the parent directory of `$path`.
//
// ```elvish-transcript
// ~> path:dir ~/bin
// ▶ /home/user
// ```

//elvdoc:fn ext
//
// ```elvish
// ext $path
// ```
//
// Outputs the extension of `$path`.
//
// ```elvish-transcript
// ~> path:ext hello.elv
// ▶ .elv
// ```

//elvdoc:fn eval-symlinks
//
// ```elvish
// eval-symlinks $path
// ```
//
// Output the path name of $path after the evaluation of symbolic links.

var Ns = eval.Ns{}.AddGoFns("path:", fns)

var fns = map[string]interface{}{
	"abs":   filepath.Abs,
	"base":  filepath.Base,
	"clean": filepath.Clean,
	"dir":   filepath.Dir,
	"ext":   filepath.Ext,

	"eval-symlinks": filepath.EvalSymlinks,
}
