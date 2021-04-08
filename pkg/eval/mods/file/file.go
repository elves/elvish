package file

import (
	"os"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

// elvdoc:fn open
// ```elvish
// file:open $filename
// ```
// Opens the file for reading.
// File should be closed after opening using close
// Example
//  ```elvish-transcript
// ~> cat a.txt
// Some text in file
// f = (file:open a.txt)
// ~> cat < $f
// Some text in file
// ~> file:close $f
// ```

func open(name string) (vals.File, error) {
	return os.Open(name)
}

// elvdoc: fn close
// flcose is used to close an already opened file.
// files are opened for reading using open
// Will not accept string name of the file
// ```elvish
// ~> file:close $fileptr
// ```

func close(f vals.File) error {
	return f.Close()
}

var Ns = eval.NsBuilder{}.AddGoFns("file:", fns).Ns()

var fns = map[string]interface{}{
	"close": close,
	"open":  open,
}
