// File module which is an equivalent of POSIX like commands.
// Checkout issule number 1263 for more information.

package file

import (
	"os"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
)

// elvdoc:fn fopen
// ```elvish
// fopen $filename
// ```
// Opens the file for reading.
// File should be closed after opening using fclose
// Example
//  ```elvish-transcript
// ~> cat a.txt
// Some text in file
// f = (fopen a.txt)
// ~> cat < $f
// Some text in file
// ~> fclose $f
// ```
func fopen(name string) (vals.File, error) {
	return os.Open(name)
}

// elvdoc: fn fclose
// flcose is used to close an already opened file.
// files are opened for reading using fopen
// Will not accept string name of the file
// ```elvish
// fclose $fileptr
// ```
func fclose(f vals.File) error {
	return f.Close()
}

//elvdoc:fn pipe
//
// ```elvish
// pipe
// ```
//
// Create a new Unix pipe that can be used in redirections.
//
// A pipe contains both the read FD and the write FD. When redirecting command
// input to a pipe with `<`, the read FD is used. When redirecting command output
// to a pipe with `>`, the write FD is used. It is not supported to redirect both
// input and output with `<>` to a pipe.
//
// Pipes have an OS-dependent buffer, so writing to a pipe without an active reader
// does not necessarily block. Pipes **must** be explicitly closed with `prclose`
// and `pwclose`.
//
// Putting values into pipes will cause those values to be discarded.
//
// Examples (assuming the pipe has a large enough buffer):
//
// ```elvish-transcript
// ~> p = (pipe)
// ~> echo 'lorem ipsum' > $p
// ~> head -n1 < $p
// lorem ipsum
// ~> put 'lorem ipsum' > $p
// ~> head -n1 < $p
// # blocks
// # $p should be closed with prclose and pwclose afterwards
// ```
//
// @cf prclose pwclose
func pipe() (vals.Pipe, error) {
	r, w, err := os.Pipe()
	return vals.NewPipe(r, w), err
}

//elvdoc:fn prclose
//
// ```elvish
// prclose $pipe
// ```
//
// Close the read end of a pipe.
//
// @cf pwclose pipe
func prclose(p vals.Pipe) error {
	return p.ReadEnd.Close()
}

var Ns = eval.NsBuilder{}.AddGoFns("file:", fns).Ns()

var fns = map[string]interface{}{
	"fclose":  fclose,
	"fopen":   fopen,
	"pipe":    pipe,
	"prclose": prclose,
}
