package clicore

import (
	"github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/newedit/editutil"
)

// Returns the first non-nil value. If all are nil, return utils.BasicMode{}
func getMode(modes ...clitypes.Mode) clitypes.Mode {
	for _, mode := range modes {
		if mode != nil {
			return mode
		}
	}
	return editutil.BasicMode{}
}
