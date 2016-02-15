package eval

import "github.com/elves/elvish/errutil"

func throw(e error) {
	errutil.Throw(e)
}

func maybeThrow(err error) {
	if err != nil {
		throw(err)
	}
}
