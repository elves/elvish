package eval

import (
	"os"
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func throw(e error) {
	util.Throw(e)
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

// parseVariable parses a variable name.
func parseVariable(qname string) (splice bool, ns string, name string) {
	if strings.HasPrefix(qname, "@") {
		splice = true
		qname = qname[1:]
		if qname == "" {
			qname = "args"
		}
	}

	i := strings.IndexRune(qname, ':')
	if i == -1 {
		return splice, "", qname
	}
	return splice, qname[:i], qname[i+1:]
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
