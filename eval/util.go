package eval

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

// ParseAndFixVariable parses a variable name. It "fixes" $@ to $@args.
func ParseAndFixVariable(qname string) (splice bool, ns string, name string) {
	splice, ns, name = ParseVariable(qname)
	if splice && ns == "" && name == "" {
		name = "args"
	}
	return splice, ns, name
}

func ParseVariable(text string) (splice bool, ns string, name string) {
	splicePart, qname := ParseVariableSplice(text)
	nsPart, name := ParseVariableQName(qname)
	ns = nsPart
	if len(ns) > 0 {
		ns = ns[:len(ns)-1]
	}
	return splicePart != "", ns, name
}

func ParseVariableSplice(text string) (splice, qname string) {
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

func MakeVariableName(splice bool, ns string, name string) string {
	prefix := ""
	if splice {
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
		// XXX should report parser bug
		panic("bad RedirMode; parser bug")
	}
}

var (
	ErrNoArgAccepted = errors.New("no argument accepted")
	ErrNoOptAccepted = errors.New("no option accepted")
)

func TakeNoArg(args []Value) {
	if len(args) > 0 {
		throw(ErrNoArgAccepted)
	}
}

func TakeNoOpt(opts map[string]Value) {
	if len(opts) > 0 {
		throw(ErrNoOptAccepted)
	}
}
