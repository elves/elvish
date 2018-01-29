package eval

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func throw(e error) {
	util.Throw(e)
}

func throwf(format string, args ...interface{}) {
	util.Throw(fmt.Errorf(format, args...))
}

func maybeThrow(err error) {
	if err != nil {
		util.Throw(err)
	}
}

func mustGetHome(uname string) string {
	dir, err := util.GetHome(uname)
	if err != nil {
		throw(err)
	}
	return dir
}

// ParseVariable parses a variable reference.
func ParseVariable(text string) (explode bool, ns string, name string) {
	explodePart, nsPart, name := SplitVariable(text)
	ns = nsPart
	if len(ns) > 0 {
		ns = ns[:len(ns)-1]
	}
	return explodePart != "", ns, name
}

// SplitVariable splits a variable reference into three parts: an optional
// explode operator (either "" or "@"), a namespace part, and a name part.
func SplitVariable(text string) (explodePart, nsPart, name string) {
	if text == "" {
		return "", "", ""
	}
	e, qname := "", text
	if text[0] == '@' {
		e = "@"
		qname = text[1:]
	}
	if qname == "" {
		return e, "", ""
	}
	i := strings.LastIndexByte(qname, ':')
	return e, qname[:i+1], qname[i+1:]
}

func MakeVariableName(explode bool, ns string, name string) string {
	prefix := ""
	if explode {
		prefix = "@"
	}
	if ns != "" {
		prefix += ns + ":"
	}
	return prefix + name
}

func makeFlag(m parse.RedirMode) int {
	switch m {
	case parse.Read:
		return os.O_RDONLY
	case parse.Write:
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case parse.ReadWrite:
		return os.O_RDWR | os.O_CREATE
	case parse.Append:
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	default:
		return -1
	}
}

var (
	ErrNoArgAccepted = errors.New("no argument accepted")
	ErrNoOptAccepted = errors.New("no option accepted")
)

func TakeNoArg(args []types.Value) {
	if len(args) > 0 {
		throw(ErrNoArgAccepted)
	}
}

func TakeNoOpt(opts map[string]types.Value) {
	if len(opts) > 0 {
		throw(ErrNoOptAccepted)
	}
}
