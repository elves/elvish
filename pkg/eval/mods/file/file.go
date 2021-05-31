package file

import (
	"math/big"
	"os"
	"strconv"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

var Ns = eval.NsBuilder{}.AddGoFns("file:", fns).Ns()

var fns = map[string]interface{}{
	"close":    close,
	"open":     open,
	"pipe":     pipe,
	"prclose":  prclose,
	"pwclose":  pwclose,
	"truncate": truncate,
}

//elvdoc:fn open
//
// ```elvish
// file:open $filename
// ```
//
// Opens a file. Currently, `open` only supports opening a file for reading.
// File must be closed with `close` explicitly. Example:
//
// ```elvish-transcript
// ~> cat a.txt
// This is
// a file.
// ~> use file
// ~> f = (file:open a.txt)
// ~> cat < $f
// This is
// a file.
// ~> close $f
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

//elvdoc:fn pipe
//
// ```elvish
// file:pipe
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
// ~> p = (file:pipe)
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
// file:prclose $pipe
// ```
//
// Close the read end of a pipe.
//
// @cf pwclose pipe

func prclose(p vals.Pipe) error {
	return p.ReadEnd.Close()
}

//elvdoc:fn pwclose
//
// ```elvish
// file:pwclose $pipe
// ```
//
// Close the write end of a pipe.
//
// @cf prclose pipe

func pwclose(p vals.Pipe) error {
	return p.WriteEnd.Close()
}

//elvdoc:fn truncate
//
// ```elvish
// file:truncate $filename $size
// ```
//
// changes the size of the named file. If the file is a symbolic link, it
// changes the size of the link's target. The size must be an integer between 0
// and 2^64-1.

func truncate(name string, rawSize vals.Num) error {
	var size int64
	switch rawSize := rawSize.(type) {
	case int:
		size = int64(rawSize)
	case *big.Int:
		if rawSize.IsInt64() {
			size = rawSize.Int64()
		} else {
			return truncateSizeOutOfRange(rawSize.String())
		}
	default:
		return errs.BadValue{
			What:  "size argument to file:truncate",
			Valid: "integer", Actual: "non-integer",
		}
	}
	if size < 0 {
		return truncateSizeOutOfRange(strconv.FormatInt(size, 10))
	}
	return os.Truncate(name, size)
}

func truncateSizeOutOfRange(size string) error {
	return errs.OutOfRange{
		What:      "size argument to file:truncate",
		ValidLow:  "0",
		ValidHigh: "2^64-1",
		Actual:    size,
	}
}
