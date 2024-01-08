package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/big"
	"os"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/sys"
)

var Ns = eval.BuildNsNamed("file").
	AddGoFns(map[string]any{
		"close":       close,
		"is-tty":      isTTY,
		"open":        open,
		"open-output": openOutput,
		"pipe":        pipe,
		"seek":        seek,
		"tell":        tell,
		"truncate":    truncate,
	}).Ns()

func isTTY(fm *eval.Frame, file any) (bool, error) {
	switch file := file.(type) {
	case *os.File:
		return sys.IsATTY(file.Fd()), nil
	case int:
		return isTTYPort(fm, file), nil
	case string:
		var fd int
		if err := vals.ScanToGo(file, &fd); err != nil {
			return false, errs.BadValue{What: "argument to file:is-tty",
				Valid: "file value or numerical FD", Actual: parse.Quote(file)}
		}
		return isTTYPort(fm, fd), nil
	default:
		return false, errs.BadValue{What: "argument to file:is-tty",
			Valid: "file value or numerical FD", Actual: vals.ToString(file)}
	}
}

func isTTYPort(fm *eval.Frame, portNum int) bool {
	p := fm.Port(portNum)
	return p != nil && sys.IsATTY(p.File.Fd())
}

func open(name string) (vals.File, error) {
	return os.Open(name)
}

type openOutputOpts struct {
	AlsoInput   bool
	IfNotExists string
	IfExists    string
	CreatePerm  int
}

func (opts *openOutputOpts) SetDefaultOptions() {
	opts.IfNotExists = "create"
	opts.IfExists = "truncate"
	opts.CreatePerm = 0o644
}

var errIfNotExistsAndIfExistsBothError = errors.New("both &if-not-exists and &if-exists are error")

func openOutput(opts openOutputOpts, name string) (vals.File, error) {
	perm := opts.CreatePerm
	if perm < 0 || perm > 0o777 {
		return nil, errs.OutOfRange{What: "create-perm option",
			ValidLow: "0", ValidHigh: "0o777", Actual: fmt.Sprintf("%O", perm)}
	}

	mode := os.O_WRONLY
	if opts.AlsoInput {
		mode = os.O_RDWR
	}
	switch opts.IfNotExists {
	case "create":
		mode |= os.O_CREATE
	case "error":
		// Do nothing: not creating is the default.
	default:
		return nil, errs.BadValue{What: "if-not-exists option",
			Valid: "create or error", Actual: parse.Quote(opts.IfNotExists)}
	}
	switch opts.IfExists {
	case "truncate":
		mode |= os.O_TRUNC
	case "append":
		mode |= os.O_APPEND
	case "update":
		// Do nothing: updating in place is the default.
	case "error":
		if mode&os.O_CREATE == 0 {
			return nil, errIfNotExistsAndIfExistsBothError
		}
		mode |= os.O_EXCL
	default:
		return nil, errs.BadValue{What: "if-exists option",
			Valid: "truncate, append, update or error", Actual: parse.Quote(opts.IfExists)}
	}

	return os.OpenFile(name, mode, fs.FileMode(perm))
}

func close(f vals.File) error {
	return f.Close()
}

func pipe() (vals.Pipe, error) {
	r, w, err := os.Pipe()
	return vals.Pipe{R: r, W: w}, err
}

type seekOpts struct {
	Whence string
}

func (opts *seekOpts) SetDefaultOptions() { opts.Whence = "start" }

func seek(opts seekOpts, f vals.File, rawOffset vals.Num) error {
	offset, err := toInt64(rawOffset, "offset", math.MinInt64, "-2^64")
	if err != nil {
		return err
	}
	var whence int
	switch opts.Whence {
	case "start":
		whence = io.SeekStart
	case "current":
		whence = io.SeekCurrent
	case "end":
		whence = io.SeekEnd
	default:
		return errs.BadValue{What: "whence",
			Valid: "start, current or end", Actual: parse.Quote(opts.Whence)}
	}
	_, err = f.Seek(offset, whence)
	return err
}

func tell(f vals.File) (vals.Num, error) {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	return vals.Int64ToNum(offset), nil
}

func truncate(name string, rawSize vals.Num) error {
	size, err := toInt64(rawSize, "size", 0, "0")
	if err != nil {
		return err
	}
	return os.Truncate(name, size)
}

func toInt64(n vals.Num, what string, validLow int64, validLowString string) (int64, error) {
	outOfRange := func() error {
		return errs.OutOfRange{What: what,
			ValidLow: validLowString, ValidHigh: "2^64-1", Actual: vals.ToString(n)}
	}
	var i int64
	switch n := n.(type) {
	case int:
		i = int64(n)
	case *big.Int:
		if n.IsInt64() {
			i = n.Int64()
		} else {
			return 0, outOfRange()
		}
	default:
		return 0, errs.BadValue{What: what,
			Valid: "exact integer", Actual: vals.ToString(n),
		}
	}
	if i < validLow {
		return 0, outOfRange()
	}
	return i, nil
}
