package eval

import "github.com/elves/elvish/util"

func throw(e error) {
	util.Throw(e)
}

func maybeThrow(err error) {
	if err != nil {
		throw(err)
	}
}
