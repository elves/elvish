package eval

import (
	"errors"

	"github.com/elves/elvish/parse"
)

// PathList wraps a list of search paths into a readonly list-like Value.
type PathList struct {
	inner *[]string
}

func (l PathList) Kind() string {
	return "list"
}

func (l PathList) Repr() string {
	var b ListReprBuilder
	for _, v := range *l.inner {
		b.WriteElem(parse.Quote(v))
	}
	return b.String()
}

func (l PathList) IndexOne(idx Value) Value {
	// XXX copied from index.go
	i := intIndex(idx)

	if i < 0 {
		i += len(*l.inner)
	}
	if i < 0 || i >= len(*l.inner) {
		throw(ErrIndexOutOfRange)
	}
	return String((*l.inner)[i])
}

func (l PathList) IndexSet(idx, v Value) {
	throw(errors.New("assignment to $paths not implemented; assign to $env:PATH instead"))
}
