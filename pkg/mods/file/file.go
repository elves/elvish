package file

import (
	_ "embed"
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
