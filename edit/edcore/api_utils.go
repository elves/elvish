package edcore

// Trivial utilities for the elvishscript API.

import (
	"fmt"

	"github.com/elves/elvish/util"
)

func throw(e error) {
	util.Throw(e)
}

func maybeThrow(e error) {
	if e != nil {
		util.Throw(e)
	}
}

func throwf(format string, args ...interface{}) {
	util.Throw(fmt.Errorf(format, args...))
}
