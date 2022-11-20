package file

import (
	"math/big"
	"os"
	"strconv"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/sys"
)

var Ns = eval.BuildNsNamed("file").
	AddGoFns(map[string]any{
		"close":    close,
		"is-tty":   isTTY,
		"open":     open,
		"pipe":     pipe,
		"truncate": truncate,
	}).Ns()

//elvdoc:fn is-tty
//
// ```elvish
// file:is-tty $file
// ```
//
// Outputs whether `$file` is a terminal device.
//
// The `$file` can be a file object or a number. If it's a number, it's
// interpreted as a numerical file descriptor.
//
// ```elvish-transcript
// ~> var f = (file:open /dev/tty)
// ~> file:is-tty $f
// ▶ $true
// ~> file:close $f
// ~> var f = (file:open /dev/null)
// ~> file:is-tty $f
// ▶ $false
// ~> file:close $f
// ~> var p = (file:pipe)
// ~> file:is-tty $p[r]
// ▶ $false
// ~> file:is-tty $p[w]
// ▶ $false
// ~> file:close $p[r]
// ~> file:close $p[w]
// ~> file:is-tty 0
// ▶ $true
// ~> file:is-tty 1
// ▶ $true
// ~> file:is-tty 2
// ▶ $true
// ```

func isTTY(fm *eval.Frame, file any) (bool, error) {
	switch file := file.(type) {
	case *os.File:
		return sys.IsATTY(file.Fd()), nil
	case int:
		return sys.IsATTY(uintptr(file)), nil
	case string:
		var fd int
		if err := vals.ScanToGo(file, &fd); err != nil {
			return false, errs.BadValue{What: "argument to file:is-tty",
				Valid: "file value or numerical FD", Actual: parse.Quote(file)}
		}
		return sys.IsATTY(uintptr(fd)), nil
	default:
		return false, errs.BadValue{What: "argument to file:is-tty",
			Valid: "file value or numerical FD", Actual: vals.ToString(file)}
	}
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
// ~> var f = (file:open a.txt)
// ~> cat < $f
// This is
// a file.
// ~> close $f
// ```
//
// @cf file:close

func open(name string) (vals.File, error) {
	return os.Open(name)
}

//elvdoc:fn close
//
// ```elvish
// file:close $file
// ```
//
// Closes a file opened with `open`.
//
// @cf file:open

func close(f vals.File) error {
	return f.Close()
}

//elvdoc:fn pipe
//
// ```elvish
// file:pipe
// ```
//
// Create a new pipe that can be used in redirections. A pipe contains a read-end and write-end.
// Each pipe object is a [pseudo-map](language.html#pseudo-map) with fields `r` (the read-end [file
// object](./language.html#file)) and `w` (the write-end).
//
// When redirecting command input from a pipe with `<`, the read-end is used. When redirecting
// command output to a pipe with `>`, the write-end is used. Redirecting both input and output with
// `<>` to a pipe is not supported.
//
// Pipes have an OS-dependent buffer, so writing to a pipe without an active reader
// does not necessarily block. Pipes **must** be explicitly closed with `file:close`.
//
// Putting values into pipes will cause those values to be discarded.
//
// Examples (assuming the pipe has a large enough buffer):
//
// ```elvish-transcript
// ~> var p = (file:pipe)
// ~> echo 'lorem ipsum' > $p
// ~> head -n1 < $p
// lorem ipsum
// ~> put 'lorem ipsum' > $p
// ~> file:close $p[w] # close the write-end
// ~> head -n1 < $p # blocks unless the write-end is closed
// ~> file:close $p[r] # close the read-end
// ```
//
// @cf file:close

func pipe() (vals.Pipe, error) {
	r, w, err := os.Pipe()
	return vals.NewPipe(r, w), err
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
