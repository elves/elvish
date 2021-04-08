package file

import (
	"os"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

var Ns = eval.NsBuilder{}.AddGoFns("file:", fns).Ns()

var fns = map[string]interface{}{
	"close": close,
	"open":  open,
}

//elvdoc:fn open
//
// ```elvish
// file:open $filename
// ```
//
// Opens a file. Currently, `fopen` only supports opening a file for reading.
// File must be closed with `fclose` explicitly. Example:
//
// ```elvish-transcript
// ~> cat a.txt
// This is
// a file.
// ~> f = (fopen a.txt)
// ~> cat < $f
// This is
// a file.
// ~> fclose $f
// ```
//
// @cf close

func open(name string) (vals.File, error) {
	return os.Open(name)
}

//elvdoc:fn close
//
// ```elvish
// ~> file:close $file
// ```
//
// Closes a file opened with `open`.
//
// @cf open

func close(f vals.File) error {
	return f.Close()
}
