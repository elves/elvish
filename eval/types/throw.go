package types

import (
	"fmt"

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
