package file

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
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
		"close":       close,
		"is-tty":      isTTY,
		"open":        open,
		"open-output": openOutput,
		"pipe":        pipe,
		"truncate":    truncate,
	}).Ns()

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

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
