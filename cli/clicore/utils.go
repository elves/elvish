package clicore

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/cliutil"
)

// Returns the first non-nil value. If all are nil, return utils.BasicMode{}
func getMode(modes ...clitypes.Mode) clitypes.Mode {
	for _, mode := range modes {
		if mode != nil {
			return mode
		}
	}
	return cliutil.BasicMode{}
}
