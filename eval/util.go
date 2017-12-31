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

func ParseVariable(text string) (explode bool, ns string, name string) {
	explodePart, qname := ParseVariableSplice(text)
	nsPart, name := ParseVariableQName(qname)
	ns = nsPart
	if len(ns) > 0 {
		ns = ns[:len(ns)-1]
	}
	return explodePart != "", ns, name
}

func ParseVariableSplice(text string) (explode, qname string) {
	if strings.HasPrefix(text, "@") {
		return "@", text[1:]
	}
	return "", text
}

func ParseVariableQName(qname string) (ns, name string) {
	i := strings.LastIndexByte(qname, ':')
	if i == -1 {
		return "", qname
	}
	return qname[:i+1], qname[i+1:]
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
